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
    id("org.curioswitch.gradle-protobuf-plugin")
    id("io.spring.dependency-management")
    id("com.moowork.node") version "1.2.0"
}

val GRPC_TOOLS_VERSION = "1.6.6"
val GRPC_TOOLS_TS_VERSION = "2.3.1"

val GRPC_VERSION = "1.14.1"
val GOOGLE_PROTOBUF_VERSION = "3.6.1"

repositories {
    jcenter()
    mavenCentral()
}

dependencies {
    protobuf(project(":api"))
}

protobuf {
    protoc {
        artifact.set("com.google.protobuf:protoc:$GOOGLE_PROTOBUF_VERSION")
    }

    // Don"t use descriptor set.
    descriptorSetOptions.path.set(file("build/descriptor"))

    languages {
        register("js") {
            outputDir.set(file("build/web"))
            option("import_style=commonjs,binary")
        }
        register("grpc_node") {
            outputDir.set(file("build/web"))
            plugin {
                path.set(file("node_modules/grpc-tools/bin/grpc_node_plugin"))
            }
        }
        register("grpc_ts") {
            outputDir.set(file("build/web"))
            plugin {
                path.set(file("node_modules/.bin/protoc-gen-ts"))
            }
        }
    }
}

val conanDir = "build/conan"
val reference = "stellarstation/stable"

val isSnapshot = (version as String).endsWith("SNAPSHOT")

tasks {
    val installProtocPlugins by registering(org.curioswitch.gradle.plugins.nodejs.tasks.NodeTask::class) {
        setCommand("npm")
        args("install", "--no-save", "grpc-tools@${GRPC_TOOLS_VERSION}", "grpc_tools_node_protoc_ts@${GRPC_TOOLS_TS_VERSION}")

        inputs.property("grpc-tools-version", GRPC_TOOLS_VERSION)
        inputs.property("grpc-tools-ts-version", GRPC_TOOLS_TS_VERSION)

        outputs.dir("node_modules/grpc-tools")
        outputs.dir("node_modules/grpc_tools_node_protoc_ts")
    }

    val generatePackageJson by registering() {
        inputs.property("version", version)
        inputs.property("grpc-tools-version", GRPC_TOOLS_VERSION)
        inputs.property("grpc-tools-ts-version", GRPC_TOOLS_TS_VERSION)
        inputs.files("package-template.json")
        outputs.file("build/web/package.json")

        doFirst {
            var packageTemplate = file("package-template.json").readText()
            packageTemplate = packageTemplate.replace("|GRPC_VERSION|", GRPC_VERSION)
                    .replace("|GOOGLE_PROTOBUF_VERSION|", GOOGLE_PROTOBUF_VERSION)
                    .replace("|API_VERSION|", version.toString())
            file("build/web/package.json").writeText(packageTemplate)
        }
    }

    val generateProto by getting(org.curioswitch.gradle.protobuf.tasks.GenerateProtoTask::class) {
        dependsOn(installProtocPlugins, generatePackageJson)

        execOverride {
            org.curioswitch.gradle.tooldownloader.DownloadedToolManager.get(project).addAllToPath(this)
        }
    }

    named("assemble").configure {
        dependsOn(generateProto)
    }

    val publish by registering(org.curioswitch.gradle.plugins.nodejs.tasks.NodeTask::class) {
        doFirst {
            copy {
                from("publishing.npmrc")
                into("build/web/")
                rename("publishing.npmrc", ".npmrc")
            }
        }

        dependsOn("build")
        args("publish", "--access=public")
        execOverride {
            environment("NPM_TOKEN", rootProject.findProperty("npm.key"))
            workingDir(file("build/web"))
        }

        onlyIf {
            !(version as String).endsWith("SNAPSHOT")
        }
    }

    named("clean").configure {
        delete(file("node_modules"))
    }
}

node {
    version = "8.11.4"
    yarnVersion = "1.9.4"
    download = true
}
