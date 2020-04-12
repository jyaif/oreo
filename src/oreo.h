#ifndef OREO_SRC_OREO_H_
#define OREO_SRC_OREO_H_

#include <cstdint>
#include <cstring>
#include <string>
#include <type_traits>
#include <utility>
#include <vector>

namespace oreo {

class SerializationArchive {
 public:
  SerializationArchive() {}

  SerializationArchive(std::vector<uint8_t> const& buffer) : buffer_(buffer) {}

  template <class... Types>
  inline void operator()(Types&&... args) {
    process(std::forward<Types>(args)...);
  }

  template <class T>
  inline void process(T&& head) {
    processImpl(head);
  }

  // Unwinds to process all data
  template <class T, class... Other>
  inline void process(T&& head, Other&&... tail) {
    process(std::forward<T>(head));
    process(std::forward<Other>(tail)...);
  }

  // For integral types and enums
  template <typename T>
  typename std::enable_if<(std::is_integral<T>::value ||
                           std::is_enum<T>::value) &&
                          !std::is_same<T, bool>::value>::type
  processImpl(T i) {
    static_assert(sizeof(T) <= 8);
    // If 2 bytes or more, use variable length integer encoding.
    if (sizeof(i) >= 2) {
      typename std::make_unsigned<T>::type unsigned_i = i;
      // Write out 7 bits at a time.
      // The high bit of the byte, when on, tells reader to continue reading
      // more bytes.
      while (unsigned_i >= 0b10000000) {
        buffer_.push_back((uint8_t)(unsigned_i | 0b10000000));
        unsigned_i = (T)(unsigned_i >> 7);
      }
      buffer_.push_back((uint8_t)unsigned_i);
    } else {
      buffer_.insert(buffer_.end(), ((uint8_t*)&i), ((uint8_t*)&i) + sizeof(i));
    }
  }

  // For booleans
  void processImpl(bool b) { buffer_.push_back(b ? 1 : 0); }

  // For strings
  void processImpl(std::string const& s) {
    uint32_t length = s.length();
    processImpl(length);
    buffer_.insert(buffer_.end(), s.begin(), s.end());
  }

  // For vectors
  template <typename T>
  void processImpl(std::vector<T> const& v) {
    uint32_t length = v.size();
    processImpl(length);
    if (sizeof(T) == 1) {
      // Speed optimisation for vectors of uint8_t and int8_t
      uint8_t* ptr = (uint8_t*)v.data();
      buffer_.insert(buffer_.end(), ptr, ptr + length);
    } else {
      for (uint32_t i = 0; i < length; i++) {
        processImpl(v[i]);
      }
    }
  }

  // For everything else
  template <typename T>
  typename std::enable_if<!(std::is_integral<T>::value ||
                            std::is_enum<T>::value)>::type
  processImpl(T a) {
    a.runArchive(*this);
  }

  std::vector<uint8_t> buffer_;
};

class DeserializationArchive {
 public:
  DeserializationArchive(void const* data) : current_cursor_(data){};

  DeserializationArchive(std::vector<uint8_t> const& buffer_)
      : DeserializationArchive(buffer_.data()) {}

  DeserializationArchive(std::vector<int8_t> const& buffer_)
      : DeserializationArchive(buffer_.data()) {}

  template <class... Types>
  inline void operator()(Types&&... args) {
    process(std::forward<Types>(args)...);
  }

  template <class T>
  inline void process(T&& head) {
    processImpl(head);
  }

  // Unwinds to process all data
  template <class T, class... Other>
  inline void process(T&& head, Other&&... tail) {
    process(std::forward<T>(head));
    process(std::forward<Other>(tail)...);
  }

  // For integral types and enums
  template <typename T>
  typename std::enable_if<(std::is_integral<T>::value ||
                           std::is_enum<T>::value) &&
                          !std::is_same<T, bool>::value>::type
  processImpl(T& i) {
    if constexpr (sizeof(i) >= 2) {
      // Read 7 bits at a time.
      // The high bit of the byte when on means to continue reading more bytes.
      typename std::make_unsigned<T>::type unsigned_i = 0;

      int shift = 0;
      uint8_t byte;
      do {
        // TODO: Check for a corrupted stream.
        // Read a nax if 3 bytes for 16 ints.
        // Read a max of 5 bytes for 32 ints.
        // Read a max of 9 bytes for 64 ints.
        // for example, for 32 ints:
        // if (shift == 5 * 7)  return false;

        // ReadByte handles end of stream cases for us.
        byte = *static_cast<const uint8_t*>(current_cursor_);
        current_cursor_ = static_cast<const char*>(current_cursor_) + 1;
        unsigned_i |= (static_cast<T>(byte & 0b1111111) << shift);
        shift += 7;
      } while ((byte & 0b10000000) != 0);
      i = (T)unsigned_i;

    } else {
      memcpy(&i, current_cursor_, sizeof(T));
      current_cursor_ = static_cast<const char*>(current_cursor_) + sizeof(T);
    }
  }

  // For booleans
  void processImpl(bool& b) {
    auto v = *static_cast<const uint8_t*>(current_cursor_);
    if (v == 0) {
      b = false;
    } else {
      b = true;
    }
    current_cursor_ = static_cast<const char*>(current_cursor_) + 1;
  }

  // For strings
  void processImpl(std::string& s) {
    uint32_t length;
    processImpl(length);
    const char* ptr = (const char*)current_cursor_;
    size_t l = length;
    s = std::string(ptr, l);
    current_cursor_ = static_cast<const char*>(current_cursor_) + length;
  }

  // For vectors
  template <typename T>
  void processImpl(std::vector<T>& v) {
    uint32_t length;
    processImpl(length);
    if (sizeof(T) == 1) {
      // Speed optimisation for vectors of uint8_t and int8_t
      T* ptr = (T*)current_cursor_;
      v.insert(v.end(), ptr, ptr + length);
      current_cursor_ = static_cast<const char*>(current_cursor_) + length;
    } else {
      v.resize(length);
      for (uint32_t i = 0; i < length; i++) {
        processImpl(v[i]);
      }
    }
  }

  // For everything else
  template <typename T>
  typename std::enable_if<!(std::is_integral<T>::value ||
                            std::is_enum<T>::value)>::type
  processImpl(T& a) {
    a.runArchive(*this);
  }

  void const* current_cursor_;
};

}  // namespace oreo

#endif  // OREO_SRC_OREO_H_
