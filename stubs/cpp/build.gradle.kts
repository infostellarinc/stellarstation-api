/*
 * Copyright 2018 Infostellar, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

plugins {
    java
    id("org.curioswitch.gradle-protobuf-plugin")
}

repositories {
    jcenter()
    mavenCentral()
}

dependencies {
    protobuf(project(":api"))
}

protobuf {
    protoc {
        path.set(file("build/conan/bin/protoc"))
    }

    // Don't use descriptor set.
    descriptorSetOptions.path.set(file("build/descriptor"))

    languages {
        register("cpp") {
            outputDir.set(file("build/conan/src"))
        }
        register("grpc") {
            outputDir.set(file("build/conan/src"))
            plugin {
                path.set(file("build/conan/bin/grpc_cpp_plugin"))
            }
        }
    }
}

val conanDir = "build/conan"
val reference = "stellarstation/stable"

tasks {
    val prepareConanFilePy by registering() {
        inputs.property("version", version)
        inputs.file("package/conanfile.py.tmpl")
        outputs.file("$buildDir/generated/scripts/conanfile.py")

        doFirst {
            val template = file("package/conanfile.py.tmpl").readText()
            val filled = template.replace("|API_VERSION|", version.toString())
            val outDir = "$buildDir/generated/scripts"
            mkdir(outDir)
            val out = file("${outDir}/conanfile.py")
            out.writeText(filled)
        }
    }

    val downloadConanDeps by registering() {
        dependsOn(prepareConanFilePy, ":toolsSetupMiniconda2Build")

        inputs.file("$buildDir/generated/scripts/conanfile.py")
        inputs.dir("package/src")
        outputs.dir("build/conan")

        val conanDir = file("build/conan")
        doFirst {
            conanDir.mkdirs()
            file("${conanDir}/src").mkdirs()

            copy {
                from("$buildDir/generated/scripts/conanfile.py")
                into(conanDir)
            }

            copy {
                from("package/src")
                into("$conanDir/src")
            }

            exec {
                commandLine("conan remote add -f bincrafters https://api.bintray.com/conan/bincrafters/public-conan && " +
                        "conan remote add -f stellarstation https://api.bintray.com/conan/infostellarinc/stellarstation-conan &&" +
                        "conan remote add -f inexorgame https://api.bintray.com/conan/inexorgame/inexor-conan && " +
                        "conan install . -s compiler=gcc -s compiler.libcxx=libstdc++ -s compiler.version=7 --build=missing --build=c-ares")
                workingDir(conanDir)

                org.curioswitch.gradle.conda.exec.CondaExecUtil.condaExec(this, project)
            }
        }
    }

    val generateProto = named("generateProto")

    generateProto.configure {
        dependsOn(downloadConanDeps)
        doLast {
            delete("build/conan/src/google")
        }
    }

    named("assemble").configure {
        dependsOn(generateProto)
    }

    val createPackage by registering() {
        dependsOn(generateProto)

        doFirst {
            exec {
                commandLine("conan create . -s compiler=gcc -s compiler.libcxx=libstdc++ -s compiler.version=7 $reference")
                workingDir(conanDir)

                org.curioswitch.gradle.conda.exec.CondaExecUtil.condaExec(this, project)
            }
        }
    }

    val uploadPackage by registering() {
        dependsOn(createPackage)

        onlyIf {
            !(version as String).endsWith("SNAPSHOT")
        }

        doFirst {
            exec {
                commandLine("conan upload stellarstation-api/$version@$reference --all -r=stellarstation")
                workingDir(conanDir)

                org.curioswitch.gradle.conda.exec.CondaExecUtil.condaExec(this, project)
            }
        }
    }

    named("compileJava") {
        enabled = false
    }

    named("compileTestJava") {
        enabled = false
    }
}
