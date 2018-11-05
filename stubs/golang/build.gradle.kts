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
    id("org.curioswitch.gradle-golang-plugin")
    id("org.curioswitch.gradle-protobuf-plugin")
    id("io.spring.dependency-management")
    id("org.ajoberstar.git-publish")
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
        artifact.set("com.google.protobuf:protoc:3.6.1")
    }

    // Don"t use descriptor set.
    descriptorSetOptions.path.set(file("build/descriptor"))

    languages {
        register("go") {
            option("plugins=grpc")
            plugin {
                path.set(file("${extra["gopath"]}/bin/protoc-gen-go"))
            }
        }
    }
}


gitPublish {
    repoUri.set("git@github.com:infostellarinc/go-stellarstation.git")
    branch.set("master")

    preserve {
        include("**")
        exclude("api")
    }

    // what to publish, this is a standard CopySpec
    contents {
        from("build/generated/proto/main/github.com/infostellarinc/go-stellarstation/api")
        into("api")
    }

    commitMessage.set("Refresh API stubs.")
}

tasks {
    val installProtocGoPlugin by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        args("install", "github.com/golang/protobuf/protoc-gen-go")
    }

    val installProtoWrap by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        args("install", "github.com/square/goprotowrap/cmd/protowrap")
    }

    val generateProto by getting(org.curioswitch.gradle.protobuf.tasks.GenerateProtoTask::class) {
        dependsOn(installProtocGoPlugin, installProtoWrap)

        execOverride {
            val protowrapPath = project.file("${extra["gopath"]}/bin/protowrap")
            setCommandLine(listOf(protowrapPath.getAbsolutePath(), "--protoc_command=${executable}") + args)

            org.curioswitch.gradle.tooldownloader.DownloadedToolManager.get(project).addAllToPath(this)
        }
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
