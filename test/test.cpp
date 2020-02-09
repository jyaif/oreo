#include <cassert>
#include <cstdio>
#include <cstdlib>
#include <string>

#include "oreo.h"

struct Bar {
  std::string a_;
  uint8_t b_;
  template <class Archive>
  void runArchive(Archive& archive) {
    archive(a_, b_);
  }
};

struct Foo {
  int8_t a_;
  uint32_t b_;
  std::string c_;
  std::vector<Bar> d_;
  template <class Archive>
  void runArchive(Archive& archive) {
    archive(a_, b_, c_, d_);
  }
};

int main() {
  Bar b0{"xyz", 19};
  Bar b1{"foo", 86};
  Foo foo0{'X', 0xdeadbeef, "abc", {b0, b1}};
  std::vector<uint8_t> expected_output = {
      'X', 0xef, 0xbe, 0xad, 0xde, 3, 0,   0,   0,   'a', 'b',
      'c', 2,    0,    0,    0,    3, 0,   0,   0,   'x', 'y',
      'z', 19,   3,    0,    0,    0, 'f', 'o', 'o', 86};

  // Test serialization
  oreo::SerializationArchive sa;
  foo0.runArchive(sa);
  assert(expected_output.size() == sa.buffer_.size());
  for (int i = 0; i < expected_output.size(); i++) {
    assert(sa.buffer_[i] == expected_output[i]);
  }
  assert(sa.buffer_ == expected_output);

  // Test deserialization
  std::vector<uint8_t> data = {3, 0, 0, 0, 'f', 'o', 'o', 86};
  oreo::DeserializationArchive da0(data);
  Bar bar{"", 0};
  bar.runArchive(da0);
  assert(bar.a_ == "foo");
  assert(bar.b_ == 86);

  // Test complex deserialization
  oreo::DeserializationArchive da1(expected_output);
  Foo foo1{0, 0, "", {}};
  foo1.runArchive(da1);
  assert(foo1.a_ == 'X');
  assert(foo1.b_ == 0xdeadbeef);
  assert(foo1.c_ == "abc");
  assert(foo1.d_.size() == 2);
  assert(foo1.d_[0].a_ == "xyz");
  assert(foo1.d_[0].b_ == 19);
  assert(foo1.d_[1].a_ == "foo");
  assert(foo1.d_[1].b_ == 86);
  printf("tests successfully passed\n");
  return EXIT_SUCCESS;
}
