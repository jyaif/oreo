#include <array>
#include <cassert>
#include <cstdio>
#include <cstdlib>
#include <string>

#include "oreo.h"

struct Bar {
  std::string a_;
  uint8_t b_;
  template <class Archive>
  bool RunArchive(Archive& archive) {
    return archive.Process(a_, b_);
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
  float h_;
  std::unique_ptr<uint32_t> i_ = std::make_unique<uint32_t>(66);
  std::unique_ptr<uint32_t> j_;
  template <class Archive>
  bool RunArchive(Archive& archive) {
    return archive.Process(a_, b_, c_, d_, e_, f_, g_, h_, i_, j_);
  }
};

template <class T>
void CheckCorrectness(std::vector<T> v) {
  oreo::SerializationArchive sa;
  sa.Process(v);
  oreo::DeserializationArchive da(sa.buffer_);
  std::vector<T> after_deserialization;
  assert(da.Process(after_deserialization));
  assert(v == after_deserialization);
}

template <class T, std::size_t N>
void CheckCorrectness(std::array<T, N> a) {
  oreo::SerializationArchive sa;
  sa.Process(a);
  oreo::DeserializationArchive da(sa.buffer_);
  std::array<T, N> after_deserialization;
  assert(da.Process(after_deserialization));
  assert(a == after_deserialization);
}

template <class T>
void CheckFailureToDeserialize(std::vector<uint8_t> data) {
  oreo::DeserializationArchive da(data);
  T v;
  assert(da.Process(v) == false);
}

template <class T>
void RemoveLastByteAndCheckFailureToDeserialize(T const& object) {
  {
    oreo::SerializationArchive sa;
    sa.Process(object);
    auto buffer_copy = sa.buffer_;
    buffer_copy.pop_back();
    CheckFailureToDeserialize<T>(buffer_copy);
  }
  // Try again with a default-initialized object of type T.
  {
    T object2;
    oreo::SerializationArchive sa;
    sa.Process(object2);
    auto buffer_copy = sa.buffer_;
    buffer_copy.pop_back();
    CheckFailureToDeserialize<T>(buffer_copy);
  }
}

template <class T>
void CheckCorrectness(T v) {
  oreo::SerializationArchive sa;
  sa.Process(v);
  oreo::DeserializationArchive da(sa.buffer_);
  T v2;
  assert(da.Process(v2));
  assert(v == v2);
  //  RemoveLastByteAndCheckFailureToDeserialize(v);
}

void CheckVarLengthInteger(int64_t v, std::vector<uint8_t> expected) {
  oreo::SerializationArchive sa;
  sa.Process(v);
  assert(sa.buffer_.size() == expected.size());
  for (int i = 0; i < expected.size(); i++) {
    assert(sa.buffer_[i] == expected[i]);
  }
  assert(sa.buffer_ == expected);
  oreo::DeserializationArchive da(sa.buffer_);
  int64_t v2;
  assert(da.Process(v2));
  assert(v == v2);
}

int main() {
  Bar b0{"xyz", 19};
  Bar b1{"foo", 86};
  Foo foo0{'X', 43, "abc", {b0, b1}, DEF, false, true, 1.5};
  // 1.5f is 0x3fc00000
  std::vector<uint8_t> expected_output = {
      'X', 43,  3,   'a', 'b', 'c', 2, 3, 'x', 'y',  'z',  19, 3,
      'f', 'o', 'o', 86,  1,   0,   1, 0, 0,   0xc0, 0x3f, 1, 66, 0};

  // Test serialization
  oreo::SerializationArchive sa;
  sa.Process(foo0);
  assert(expected_output.size() == sa.buffer_.size());
  for (int i = 0; i < expected_output.size(); i++) {
    assert(sa.buffer_[i] == expected_output[i]);
  }
  assert(sa.buffer_ == expected_output);

  // Test deserialization
  std::vector<uint8_t> data = {3, 'f', 'o', 'o', 86};
  oreo::DeserializationArchive da0(data);
  Bar bar{"", 0};
  assert(da0.Process(bar));
  assert(bar.a_ == "foo");
  assert(bar.b_ == 86);

  // Test complex deserialization
  oreo::DeserializationArchive da1(expected_output);
  Foo foo1{0, 0, "", {}, ABC, false, false, 0.0f, nullptr, nullptr};
  assert(da1.Process(foo1));
  assert(foo1.a_ == 'X');
  assert(foo1.b_ == 43);
  assert(foo1.c_ == "abc");
  assert(foo1.d_.size() == 2);
  assert(foo1.d_[0].a_ == "xyz");
  assert(foo1.d_[0].b_ == 19);
  assert(foo1.d_[1].a_ == "foo");
  assert(foo1.d_[1].b_ == 86);
  assert(foo1.e_ == DEF);
  assert(foo1.f_ == false);
  assert(foo1.g_ == true);
  assert(foo1.h_ == 1.5f);
  assert(foo1.i_ != nullptr);
  assert(*foo1.i_ == 66);
  assert(foo1.j_ == nullptr);

  // Test variable length integers
  CheckVarLengthInteger(0, {0});
  CheckVarLengthInteger(1, {1});
  CheckVarLengthInteger(-1, {255, 255, 255, 255, 255, 255, 255, 255, 255, 1});
  CheckVarLengthInteger(-2, {254, 255, 255, 255, 255, 255, 255, 255, 255, 1});
  CheckVarLengthInteger(127, {127});
  CheckVarLengthInteger(128, {128, 1});
  CheckVarLengthInteger(200, {200, 1});
  CheckVarLengthInteger(255, {255, 1});
  CheckVarLengthInteger(256, {128, 2});
  CheckVarLengthInteger(300, {172, 2});
  CheckVarLengthInteger(32767, {255, 255, 1});
  CheckVarLengthInteger(32768, {128, 128, 2});
  CheckVarLengthInteger(65535, {255, 255, 3});
  CheckVarLengthInteger(65536, {128, 128, 4});
  CheckVarLengthInteger(0x7fffffff, {255, 255, 255, 255, 7});
  CheckVarLengthInteger(0x80000000, {128, 128, 128, 128, 8});
  CheckVarLengthInteger(0xffffffff, {255, 255, 255, 255, 15});
  CheckVarLengthInteger(0x111111111111111,
                        {145, 162, 196, 136, 145, 162, 196, 136, 1});
  CheckVarLengthInteger(0x7fffffffffffffff,
                        {255, 255, 255, 255, 255, 255, 255, 255, 127});
  CheckVarLengthInteger(0xffffffffffffffff,
                        {255, 255, 255, 255, 255, 255, 255, 255, 255, 1});
  const std::vector<int16_t> int16s = {
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
  const std::vector<uint16_t> uint16s = {0,   1,    2,    10,    100,   200,
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

  // Test floating point values
  CheckCorrectness(0.0f);
  CheckCorrectness(-0.0f);
  CheckCorrectness(1.6f);
  CheckCorrectness(-42.6f);

  // Test deserialization errors
  std::vector<std::vector<uint8_t>> datas;
  for (int i = 0; i < 20; i++) {
    datas.push_back({});
    for (int j = 0; j < i; j++) {
      datas[i].push_back(0xff);
    }
  }
  assert(datas[0].empty());
  assert(datas[1].size() == 1);
  assert(datas[2].size() == 2);

  {
    // integers and enums
    enum E32 : int32_t { Foo };
    for (auto& data : datas) {
      CheckFailureToDeserialize<int16_t>(data);
      CheckFailureToDeserialize<uint16_t>(data);
      CheckFailureToDeserialize<int32_t>(data);
      CheckFailureToDeserialize<uint32_t>(data);
      CheckFailureToDeserialize<int64_t>(data);
      CheckFailureToDeserialize<uint64_t>(data);
      CheckFailureToDeserialize<E32>(data);
    }
  }
  {
    // vectors
    for (auto& data : datas) {
      CheckFailureToDeserialize<std::vector<int8_t>>(data);
    }
  }
  {
    // Everything, removing one byte at the end.
    bool b = true;
    uint8_t uint8 = 54;
    uint16_t uint16 = 5432;
    uint32_t uint32 = 54320778;
    int8_t int8 = 54;
    int16_t int16 = 5432;
    int32_t int32 = 54320778;
    std::string string = "foobar";
    std::vector<uint8_t> vec_uint8 = {0, 1, 56};
    Foo foo{0, 0, "", {}, ABC};

    RemoveLastByteAndCheckFailureToDeserialize(b);
    RemoveLastByteAndCheckFailureToDeserialize(uint8);
    RemoveLastByteAndCheckFailureToDeserialize(uint16);
    RemoveLastByteAndCheckFailureToDeserialize(uint32);
    RemoveLastByteAndCheckFailureToDeserialize(int8);
    RemoveLastByteAndCheckFailureToDeserialize(int16);
    RemoveLastByteAndCheckFailureToDeserialize(int32);
    RemoveLastByteAndCheckFailureToDeserialize(string);
    RemoveLastByteAndCheckFailureToDeserialize(vec_uint8);
    RemoveLastByteAndCheckFailureToDeserialize(foo);
  }

  {
    // Test std::array
    std::array<uint8_t, 4> array0 = {0x12, 0x56, 0x34, 0xab};
    CheckCorrectness(array0);
    std::array<uint32_t, 3> array1 = {0x777777, 0xa8, 0x8d786a};
    CheckCorrectness(array1);
  }

  printf("tests successfully passed\n");
  return EXIT_SUCCESS;
}
