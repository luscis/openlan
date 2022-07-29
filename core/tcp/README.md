# build

mkdir -p build && cd ./build
cmake ..
make

# valgrind

valgrind --leak-check=yes ./build/core
