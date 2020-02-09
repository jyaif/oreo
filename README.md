# oreo

Minimalist serialization library inspired by Cereal

Some advantages compared to Cereal:
* Does not use RTTI.
* Lib is very few LOC. It's easier to customise, and with faster compile time.

Main disadvantages compared to Cereal:
* Only serializes/deserializes structs with integers, std::string, std::vector.
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
