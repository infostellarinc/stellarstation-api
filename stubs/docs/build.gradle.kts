/*
 * Copyright 2018 Infostellar, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

plugins {
    id("org.curioswitch.gradle-golang-plugin")
    id("org.curioswitch.gradle-protobuf-plugin")
}

repositories {
    jcenter()
    mavenCentral()
}

dependencies {
    protobuf(project(":api"))
}

val GOPATH = extra["gopath"]

protobuf {
    protoc {
        artifact.set("com.google.protobuf:protoc:3.6.1")
    }

    // Don"t use descriptor set.
    descriptorSetOptions.path.set(file("build/descriptor"))

    languages {
        register("doc") {
            outputDir.set(file("build/apidocs"))
            plugin {
                path.set(file("${GOPATH}/bin/protoc-gen-doc"))
            }
        }
    }
}


tasks {
    val installProtocDocPlugin by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        dependsOn(named("goDownloadDeps"))

        args("install", "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc")

        execCustomizer {
            environment("CGO_ENABLED", "0")
        }
    }

    val generateProto by getting(org.curioswitch.gradle.protobuf.tasks.GenerateProtoTask::class) {
        dependsOn(installProtocDocPlugin)
    }

    named("assemble").configure({
        dependsOn(generateProto)
    })

    // Only generated code, no need to check.
    named("goCheck").configure({
        enabled = false
    })
    named("goTest").configure({
        enabled = false
    })
}
