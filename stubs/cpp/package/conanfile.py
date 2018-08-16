from conans import ConanFile, CMake


class StellarstationapiConan(ConanFile):
    name = "stellarstation-api"
    version = "0.0.8"
    license = "Apache-2.0"
    url = "https://github.com/infostellarinc/stellarstation-api"
    description = "C++ gRPC stubs for conneting to the StellarStation API."
    settings = "os", "compiler", "build_type", "arch"
    options = {"shared": [True, False]}
    default_options = "shared=False"
    generators = "cmake"
    exports_sources = "src/*"
    requires = "grpc/1.14.1@inexorgame/stable", "protobuf/3.6.1@bincrafters/stable"

    def build(self):
        cmake = CMake(self)
        cmake.configure(source_folder="src")
        cmake.build()

    def imports(self):
        self.copy("grpc_cpp_plugin", "bin", "bin", 'grpc')
        self.copy("protoc", "bin", "bin", 'protobuf')

    def package(self):
        self.copy("*.h", dst="include", src="src")
        self.copy("*.lib", dst="lib", keep_path=False)
        self.copy("*.dll", dst="bin", keep_path=False)
        self.copy("*.dylib*", dst="lib", keep_path=False)
        self.copy("*.so", dst="lib", keep_path=False)
        self.copy("*.a", dst="lib", keep_path=False)

    def package_info(self):
        self.cpp_info.libs = ["hello"]
