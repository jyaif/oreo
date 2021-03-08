# oreo

[![Build Status](https://api.travis-ci.org/jyaif/oreo.svg)](https://travis-ci.org/jyaif/oreo)

Minimalist serialization library inspired by [Cereal](https://github.com/USCiLab/cereal).
It's used by a game for serializing network messages.

Some disadvantages compared to Cereal:
* Only serializes/deserializes structs with booleans, integers, floats, enums, std::string, std::vector, std::array.
* Only serializes/deserializes to binary, with the endianness of the system.
* No documentation and no efforts made to give useful compile-time error messages.
* No versioning.

Some advantages compared to Cereal:
* Does not use RTTI.
* Lib is very few LOC: it is easier to customise, and has faster compile time.
* Smaller footprint in the final binary. I shaved 8kb from a binary in release mode by switching ~20 smallish classes from cereal to oreo.

TODO:
* Allow configuring maximum string/vector size when deserializing.

---

To run tests locally:

```
mkdir -p out
cd out
cmake ..
make
./oreo_test_bin
```

for Xcode:
```
mkdir -p out2
cd out2
cmake -G "Xcode" ..
```
