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

enum QuxEnum : int8_t { ABC, DEF };

struct Foo {
  int8_t a_;
  uint32_t b_;
  std::string c_;
  std::vector<Bar> d_;
  QuxEnum e_;
  bool f_;
  bool g_;
  template <class Archive>
  void runArchive(Archive& archive) {
    archive(a_, b_, c_, d_, e_, f_, g_);
  }
};

template <class T>
void CheckCorrectness(std::vector<T> v) {
  oreo::SerializationArchive sa;
  sa.process(v);
  oreo::DeserializationArchive da(sa.buffer_);
  std::vector<T> after_deserialization;
  da.process(after_deserialization);
  assert(v == after_deserialization);
}

int main() {
  Bar b0{"xyz", 19};
  Bar b1{"foo", 86};
  Foo foo0{'X', 43, "abc", {b0, b1}, DEF, false, true};
  std::vector<uint8_t> expected_output = {'X', 43,  3,   'a', 'b', 'c', 2,
                                          3,   'x', 'y', 'z', 19,  3,   'f',
                                          'o', 'o', 86,  1,   0,   1};

  // Test serialization
  oreo::SerializationArchive sa;
  foo0.runArchive(sa);
  assert(expected_output.size() == sa.buffer_.size());
  for (int i = 0; i < expected_output.size(); i++) {
    assert(sa.buffer_[i] == expected_output[i]);
  }
  assert(sa.buffer_ == expected_output);

  // Test deserialization
  std::vector<uint8_t> data = {3, 'f', 'o', 'o', 86};
  oreo::DeserializationArchive da0(data);
  Bar bar{"", 0};
  bar.runArchive(da0);
  assert(bar.a_ == "foo");
  assert(bar.b_ == 86);

  // Test complex deserialization
  oreo::DeserializationArchive da1(expected_output);
  Foo foo1{0, 0, "", {}, ABC};
  foo1.runArchive(da1);
  assert(foo1.a_ == 'X');
  assert(foo1.b_ == 43);
  assert(foo1.c_ == "abc");
  assert(foo1.d_.size() == 2);
  assert(foo1.d_[0].a_ == "xyz");
  assert(foo1.d_[0].b_ == 19);
  assert(foo1.d_[1].a_ == "foo");
  assert(foo1.d_[1].b_ == 86);
  assert(foo1.e_ == DEF);

  // Test variable length integers
  const std::vector<int32_t> int16s = {
      0,  1,  2,   10,   100,  200,  300,   1000,  5000,   10000, 32767,
      -1, -2, -10, -100, -200, -300, -1000, -5000, -10000, -32768};
  CheckCorrectness(int16s);
  const std::vector<int32_t> int32s = {
      0,    1,      2,      10,       100,        200,     300,     1000,
      5000, 100000, 400000, 50000000, 0x7fffffff, -1,      -2,      -10,
      -100, -200,   -300,   -1000,    -5000,      -100000, -400000, -50000000};
  CheckCorrectness(int32s);
  const std::vector<int64_t> int64s = {0,
                                       1,
                                       2,
                                       10,
                                       100,
                                       200,
                                       300,
                                       1000,
                                       5000,
                                       100000,
                                       400000,
                                       50000000,
                                       0x111111111111111,
                                       0x7fffffffffffffff,
                                       -1,
                                       -2,
                                       -10,
                                       -100,
                                       -200,
                                       -300,
                                       -1000,
                                       -5000,
                                       -100000,
                                       -400000,
                                       -50000000};
  CheckCorrectness(int64s);
  const std::vector<uint32_t> uint16s = {0,   1,    2,    10,    100,   200,
                                         300, 1000, 5000, 10000, 32767, 0xffff};
  CheckCorrectness(uint16s);
  const std::vector<uint32_t> uint32s = {
      0,    1,    2,      10,     100,      200,        300,
      1000, 5000, 100000, 400000, 50000000, 0x7fffffff, 0xffffffff};
  CheckCorrectness(uint32s);
  const std::vector<uint64_t> uint64s = {0,
                                         1,
                                         2,
                                         10,
                                         100,
                                         200,
                                         300,
                                         1000,
                                         5000,
                                         100000,
                                         400000,
                                         50000000,
                                         0x111111111111111,
                                         0x7fffffffffffffff,
                                         0xffffffffffffffff};
  CheckCorrectness(uint64s);

  printf("tests successfully passed\n");
  return EXIT_SUCCESS;
}
