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
    `java-library`
    id("org.curioswitch.gradle-protobuf-plugin")
}

val packageDir = file("build/publications/python").getAbsolutePath()

repositories {
    jcenter()
    mavenCentral()
}

dependencies {
    protobuf(project(":api"))
}

protobuf {
    protoc {
        path.set(file("$buildDir/generated/scripts/run_protoc.sh"))
    }

    // Don"t use descriptor set.
    descriptorSetOptions.path.set(file("build/descriptor"))

    languages {
        register("python") {
            outputDir.set(file(packageDir))
        }

        register("grpc_python") {
            outputDir.set(file(packageDir))
        }
    }
}

tasks {
    val fillRunProtocScript by registering() {
        doFirst {
            val template = file("src/misc/run_protoc.sh.tmpl").readText()
            val filled = template.replaceFirst(
                    "|CONDA_PROFILE_PATH|",
                    "${org.curioswitch.gradle.tooldownloader.DownloadedToolManager.get(project).getToolDir("miniconda-build")}/etc/profile.d/conda.sh")
            val outDir = "$buildDir/generated/scripts"
            mkdir(outDir)
            val out = file("${outDir}/run_protoc.sh")
            out.writeText(filled)
            out.setExecutable(true)
        }
    }

    val prepareSetupPy by registering() {
        inputs.property("version", version);
        inputs.file("src/misc/python/setup.py.tmpl")
        outputs.file("$buildDir/generated/scripts/setup.py")

        doFirst {
            val template = file("src/misc/python/setup.py.tmpl").readText()
            val filled = template.replace("|API_VERSION|", (version as String))
            val outDir = "$buildDir/generated/scripts"
            mkdir(outDir)
            val out = file("${outDir}/setup.py")
            out.writeText(filled)
            out.setExecutable(true)

            // pydocstring does not work with a default ASCII encoding so we need to force UTF-8.
            file("$buildDir/generated/sitecustomize.py").writeText("import sys; sys.setdefaultencoding('utf8')")
        }
    }

    val generateProto = named<org.curioswitch.gradle.protobuf.tasks.GenerateProtoTask>("generateProto")
    generateProto.configure {
        dependsOn(fillRunProtocScript, prepareSetupPy)

        execOverride {
            environment("PYTHONPATH", "$buildDir/generated")
        }
    }

    val setupPythonPackage by registering() {
        dependsOn(generateProto, prepareSetupPy)

        doFirst {
            mkdir(packageDir)

            copy {
                from("$buildDir/generated/scripts/setup.py")
                from(rootProject.file("README.md"))
                from(rootProject.file("LICENSE"))
                into(packageDir)
            }

            copy {
                from("$buildDir/generated/source/proto/main/python/stellarstation")
                into("$packageDir/stellarstation")
            }

            file("$packageDir/stellarstation/__init__.py").writeText("name = 'stellarstation'")

            file("$packageDir/stellarstation/api").walk()
                    .filter { it.isDirectory }
                    .forEach { file("$it/__init__.py").writeText("") }
        }
    }

    val buildPythonPackage by registering() {
        dependsOn(setupPythonPackage)

        doFirst {
            exec {
                commandLine("python setup.py sdist bdist_wheel")
                workingDir(packageDir)

                org.curioswitch.gradle.conda.exec.CondaExecUtil.condaExec(this, project)
            }
        }
    }

    val uploadPythonPackage by registering() {
        dependsOn(buildPythonPackage)

        doFirst {
            exec {
                val user = rootProject.property("pypi.user")
                val password = rootProject.property("pypi.password")

                commandLine("twine upload --skip-existing -u $user -p $password dist/*")
                workingDir(packageDir)

                org.curioswitch.gradle.conda.exec.CondaExecUtil.condaExec(this, project)
            }
        }

        onlyIf {
            !(version as String).endsWith("SNAPSHOT")
        }
    }

    val generateDocs by registering() {
        dependsOn(buildPythonPackage)

        doFirst {
            exec {
                commandLine("pdoc --html stellarstation --overwrite --template-dir=${file("src/misc/templates").getAbsolutePath()}")
                mkdir("$buildDir/docs")
                workingDir("$buildDir/docs")

                // pdoc does not work with a default ASCII encoding so we need to force UTF-8.
                file("$packageDir/build/lib/sitecustomize.py").writeText("import sys; sys.setdefaultencoding('utf8')")
                environment("PYTHONPATH", "$packageDir/build/lib")

                org.curioswitch.gradle.conda.exec.CondaExecUtil.condaExec(this, project)
            }
        }
    }

    named("assemble").configure {
        dependsOn(buildPythonPackage)
    }

    named("compileJava") {
        enabled = false
    }

    named("compileTestJava") {
        enabled = false
    }
}
