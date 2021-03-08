#ifndef OREO_SRC_OREO_H_
#define OREO_SRC_OREO_H_

#include <array>
#include <cstdint>
#include <cstring>
#include <string>
#include <type_traits>
#include <utility>
#include <vector>

namespace oreo {

// 1 GB.
constexpr size_t kMaxStringLength = 1073741824;
constexpr size_t kMaxVectorElementCount = 1073741824;

class SerializationArchive {
 public:
  SerializationArchive() {}

  SerializationArchive(std::vector<uint8_t> const& buffer) : buffer_(buffer) {}

  template <class T>
  inline bool Process(T&& head) {
    ProcessImpl(head);
    return true;
  }

  // Unwinds to process all data
  template <class T, class... Other>
  inline bool Process(T&& head, Other&&... tail) {
    ProcessImpl(std::forward<T>(head));
    Process(std::forward<Other>(tail)...);
    return true;
  }

  // For integral types and enums
  template <typename T>
  typename std::enable_if<(std::is_integral<T>::value ||
                           std::is_enum<T>::value) &&
                          !std::is_same<T, bool>::value>::type
  ProcessImpl(T i) {
    static_assert(sizeof(T) <= 8);
    // If 2 bytes or more, use variable length integer encoding.
    if constexpr (sizeof(i) >= 2) {
      typename std::make_unsigned<T>::type unsigned_i = i;
      // Write out 7 bits at a time.
      // The high bit of the byte, when on, tells reader to continue reading
      // more bytes.
      while (unsigned_i >= 0b10000000) {
        buffer_.push_back(static_cast<uint8_t>(unsigned_i | 0b10000000));
        unsigned_i = static_cast<T>(unsigned_i >> 7);
      }
      buffer_.push_back(static_cast<uint8_t>(unsigned_i));
    } else {
      buffer_.insert(buffer_.end(), reinterpret_cast<uint8_t*>(&i),
                     reinterpret_cast<uint8_t*>(&i) + sizeof(i));
    }
  }

  // For floats
  void ProcessImpl(float& float_value) {
    const uint8_t* casted_ptr = reinterpret_cast<const uint8_t*>(&float_value);
    ProcessArray(casted_ptr, sizeof(float));
  }

  // For booleans
  void ProcessImpl(bool b) { buffer_.push_back(b ? 1 : 0); }

  // For strings
  void ProcessImpl(std::string const& s) {
    uint32_t length = static_cast<uint32_t>(s.length());
    ProcessImpl(length);
    buffer_.insert(buffer_.end(), s.begin(), s.end());
  }

  // For vectors
  template <typename T>
  void ProcessImpl(std::vector<T> const& v) {
    uint32_t length = static_cast<uint32_t>(v.size());
    ProcessImpl(length);
    if constexpr (sizeof(T) == 1) {
      // Speed optimisation for vectors of uint8_t and int8_t
      const uint8_t* ptr = reinterpret_cast<const uint8_t*>(v.data());
      buffer_.insert(buffer_.end(), ptr, ptr + length);
    } else {
      for (uint32_t i = 0; i < length; i++) {
        ProcessImpl(v[i]);
      }
    }
  }

  // For std::arrays
  template <typename T, std::size_t N>
  void ProcessImpl(std::array<T, N> const& v) {
    ProcessArray(v.data(), N);
  }

  template <typename T>
  void ProcessArray(T* ptr, size_t N) {
    if constexpr (sizeof(T) == 1) {
      // Speed optimisation for vectors of uint8_t and int8_t
      const uint8_t* casted_ptr = reinterpret_cast<const uint8_t*>(ptr);
      buffer_.insert(buffer_.end(), casted_ptr, casted_ptr + N);
    } else {
      for (uint32_t i = 0; i < N; i++) {
        ProcessImpl(ptr[i]);
      }
    }
  }

  // For everything else
  template <typename T>
  typename std::enable_if<!(std::is_integral<T>::value ||
                            std::is_enum<T>::value)>::type
  ProcessImpl(T const& a) {
    const_cast<T&>(a).RunArchive(*this);
  }

  std::vector<uint8_t> buffer_;
};

class DeserializationArchive {
 public:
  // |end| is the theoretical element that would follow the last element in the
  // vector.
  DeserializationArchive(const uint8_t* data, const uint8_t* end)
      : current_cursor_(data), end_cursor_(end){};

  DeserializationArchive(std::vector<uint8_t> const& buffer)
      : DeserializationArchive(buffer.data(), &*buffer.end()) {}

  DeserializationArchive(std::vector<int8_t> const& buffer)
      : DeserializationArchive(
            reinterpret_cast<const uint8_t*>(buffer.data()),
            reinterpret_cast<const uint8_t*>(&*buffer.end())) {}

  template <class T>
  [[nodiscard]] inline bool Process(T&& head) {
    return ProcessImpl(head);
  }

  // Unwinds to process all data
  template <class T, class... Other>
  [[nodiscard]] inline bool Process(T&& head, Other&&... tail) {
    bool success = ProcessImpl(std::forward<T>(head));
    if (!success) {
      return false;
    }
    return Process(std::forward<Other>(tail)...);
  }

  // For integral types and enums
  template <typename T>
  [[nodiscard]] typename std::enable_if<(std::is_integral<T>::value ||
                                         std::is_enum<T>::value) &&
                                            !std::is_same<T, bool>::value,
                                        bool>::type
  ProcessImpl(T& i) {
    if constexpr (sizeof(T) >= 2) {
      // Read 7 bits at a time.
      // The high bit of the byte when on means to continue reading more bytes.
      typename std::make_unsigned<T>::type unsigned_i = 0;

      uint32_t shift = 0;
      uint8_t byte;
      do {
        if (current_cursor_ >= end_cursor_) {
          return false;
        }
        // Check for corrupted stream.
        // Read a max if 3 bytes for 16 ints.
        // Read a max of 5 bytes for 32 ints.
        // Read a max of 9 bytes for 64 ints.
        if (shift > (sizeof(T) + 1) * 7) {
          return false;
        }

        // ReadByte handles end of stream cases for us.
        byte = *current_cursor_;
        current_cursor_++;
        unsigned_i |= (static_cast<T>(byte & 0b1111111) << shift);
        shift += 7;
      } while ((byte & 0b10000000) != 0);
      i = (T)unsigned_i;

    } else {
      if (current_cursor_ >= end_cursor_) {
        return false;
      }
      memcpy(&i, current_cursor_, sizeof(T));
      current_cursor_ += sizeof(T);
    }
    return true;
  }

  // For floats.
  [[nodiscard]] bool ProcessImpl(float& f) {
    uint8_t* ptr = reinterpret_cast<uint8_t*>(&f);
    return ProcessImpl(ptr, sizeof(float));
  }

  // For booleans
  [[nodiscard]] bool ProcessImpl(bool& b) {
    if (current_cursor_ >= end_cursor_) {
      return false;
    }
    auto v = *current_cursor_;
    if (v == 0) {
      b = false;
    } else {
      b = true;
    }
    current_cursor_++;
    return true;
  }

  // For strings
  [[nodiscard]] bool ProcessImpl(std::string& s) {
    uint32_t length;
    auto read_length_success = ProcessImpl(length);
    if (!read_length_success) {
      return false;
    }
    if (length > kMaxStringLength) {
      return false;
    }
    if (current_cursor_ + length > end_cursor_) {
      return false;
    }
    const char* ptr = reinterpret_cast<const char*>(current_cursor_);

    size_t l = length;
    s = std::string(ptr, l);
    current_cursor_ += length;
    return true;
  }

  // For vectors
  template <typename T>
  [[nodiscard]] bool ProcessImpl(std::vector<T>& v) {
    uint32_t length;
    auto read_length_success = ProcessImpl(length);
    if (!read_length_success) {
      return false;
    }
    if (length > kMaxVectorElementCount) {
      return false;
    }
    if constexpr (sizeof(T) == 1) {
      if (current_cursor_ + length > end_cursor_) {
        return false;
      }
      // Speed optimisation for vectors of uint8_t and int8_t.
      const T* ptr = reinterpret_cast<const T*>(current_cursor_);
      v.insert(v.end(), ptr, ptr + length);
      current_cursor_ += length;
    } else {
      v.resize(length);
      for (uint32_t i = 0; i < length; i++) {
        auto success = ProcessImpl(v[i]);
        if (!success) {
          return false;
        }
      }
    }
    return true;
  }

  // For arrays
  template <typename T, std::size_t N>
  [[nodiscard]] bool ProcessImpl(std::array<T, N>& v) {
    return ProcessImpl(v.data(), N);
  }

  template <typename T>
  [[nodiscard]] bool ProcessImpl(T* dest, std::size_t N) {
    if constexpr (sizeof(T) == 1) {
      if (current_cursor_ + N > end_cursor_) {
        return false;
      }
      // Speed optimisation for vectors of uint8_t and int8_t.
      const T* ptr = reinterpret_cast<const T*>(current_cursor_);
      memcpy(dest, ptr, N);
      current_cursor_ += N;
    } else {
      for (uint32_t i = 0; i < N; i++) {
        auto success = ProcessImpl(dest[i]);
        if (!success) {
          return false;
        }
      }
    }
    return true;
  }

  // For everything else
  template <typename T>
  [[nodiscard]] typename std::enable_if<!(std::is_integral<T>::value ||
                                          std::is_enum<T>::value),
                                        bool>::type
  ProcessImpl(T& a) {
    return a.RunArchive(*this);
  }

  const uint8_t* current_cursor_;
  const uint8_t* end_cursor_;
};

}  // namespace oreo

#endif  // OREO_SRC_OREO_H_
