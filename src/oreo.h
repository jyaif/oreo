#ifndef OREO_SRC_OREO_H_
#define OREO_SRC_OREO_H_

#include <cstdint>
#include <cstring>
#include <string>
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
  typename std::enable_if<std::is_integral<T>::value ||
                          std::is_enum<T>::value>::type
  processImpl(T a) {
    buffer_.insert(buffer_.end(), ((uint8_t*)&a), ((uint8_t*)&a) + sizeof(a));
  }

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
    for (uint32_t i = 0; i < length; i++) {
      processImpl(v[i]);
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
  typename std::enable_if<std::is_integral<T>::value ||
                          std::is_enum<T>::value>::type
  processImpl(T& a) {
    memcpy(&a, current_cursor_, sizeof(T));
    current_cursor_ = static_cast<const char*>(current_cursor_) + sizeof(T);
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
    v.resize(length);
    for (uint32_t i = 0; i < length; i++) {
      processImpl(v[i]);
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
