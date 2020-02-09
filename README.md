# oreo

[![Build Status](https://api.travis-ci.org/jyaif/oreo.svg)](https://travis-ci.org/jyaif/oreo)

Minimalist serialization library inspired by [Cereal](https://github.com/USCiLab/cereal).

Some advantages compared to Cereal:
* Does not use RTTI.
* Lib is very few LOC: it is easier to customise, and has faster compile time.
* Smaller footprint in the final binary. I shaved 8kb from a binary in release mode by switching ~20 classes from cereal to oreo.

Some disadvantages compared to Cereal:
* Only serializes/deserializes structs with integers, enums, std::string, std::vector.
* Only serializes/deserializes to CPU-agnostic binary.

TODO:
* Enforce maximum string length when serializing.
* Handle errors when deserializing.
* Variable width integer

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
