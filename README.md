# oreo

> [!WARNING]
> Contributions that break change the binary representation are unlikely to be accepted.
> The binary representation of integers will change in the future.
> Not production ready. Code should be used as a basis for rolling out your own serialization.

Minimalist serialization library inspired by [Cereal](https://github.com/USCiLab/cereal).

Some disadvantages compared to Cereal:
* Only serializes/deserializes structs with booleans, integers, floats, enums, std::string, std::vector, std::array.
* Only serializes/deserializes to binary, with the endianness of the system.
* No documentation and no efforts made to give useful compile-time error messages.
* No built-in versioning.

Some advantages compared to Cereal:
* Does not use RTTI.
* Lib is very few LOC: it is easier to customise, and has faster compile time.
* Smaller footprint in the final binary. I shaved 8kb from a binary in release mode by switching ~20 smallish classes from cereal to oreo.
* A Go implementation of the library exists.

TODO:
* Allow configuring maximum string/vector size when deserializing.

---

To run the cpp tests locally:

```
cd cpp
mkdir -p out
cd out
cmake ..
make
./oreo_test_bin
```

for Xcode:
```
cd cpp
mkdir -p out2
cd out2
cmake -G "Xcode" ..
```
