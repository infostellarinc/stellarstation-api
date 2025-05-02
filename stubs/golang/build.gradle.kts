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
    id("org.ajoberstar.git-publish")
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
        artifact.set("com.google.protobuf:protoc:4.30.2")
    }

    // Don"t use descriptor set.
    descriptorSetOptions.path.set(file("build/descriptor"))

    languages {
        register("go") {
            option("plugins=grpc")
        }
    }
}


gitPublish {
    repoUri.set("git@github.com:infostellarinc/go-stellarstation.git")
    branch.set("master")

    preserve {
        include("**")
        exclude("api")
        exclude("vendor")
    }

    // what to publish, this is a standard CopySpec
    contents {
        from("build/generated/proto/main/github.com/infostellarinc/go-stellarstation/api") {
            into("api")
        }
        from("build/generated/proto/main/github.com/infostellarinc/go-stellarstation/vendor") {
            into("vendor")
        }
        from("build/generated/proto/main/github.com/infostellarinc/go-stellarstation") {
            include("go.mod", "go.sum")
        }
    }

    commitMessage.set("Refresh API stubs.")
}

tasks {
    val installProtocGoPlugin by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        dependsOn(named("goDownloadDeps"))

        args("install", "google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6")
    }

    val installProtocGoGrpcPlugin by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        dependsOn(named("goDownloadDeps"))

        args("install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1")
    }

    val copyGoMod by registering(Copy::class) {
        into("build/generated/proto/main/github.com/infostellarinc/go-stellarstation")
        from("go.mod")
        from("go.sum")
    }

    val copyBuildedSource by registering(Copy::class) {
        into("build/goproto/src/github.com/infostellarinc/go-stellarstation")
        from("build/generated/proto/main/github.com/infostellarinc/go-stellarstation")
        dirMode = 493 // 755
        fileMode = 420 // 644

        dependsOn(generateProto)
    }

    val generateProto by getting(org.curioswitch.gradle.protobuf.tasks.GenerateProtoTask::class) {
        dependsOn(installProtocGoPlugin, installProtocGoGrpcPlugin, extractProto)
        val protoDir = "${project.projectDir}/build/extracted-protos/main"
        val outputDir = "${project.projectDir}/build/generated/proto/main"

        execOverride {
            val command = mutableListOf(
                "${executable}",
                "--go_out=$outputDir",
                "--go-grpc_out=$outputDir",
                "--plugin=protoc-gen-go=${GOPATH}/bin/protoc-gen-go",
                "--plugin=protoc-gen-go-grpc=${GOPATH}/bin/protoc-gen-go-grpc",
                "-I=$protoDir"
            )

            // Add all proto files from fileTree
            command.addAll(fileTree(protoDir).matching { include("**/*.proto") }.files.map { it.absolutePath })

            setCommandLine(command)
            org.curioswitch.gradle.tooldownloader.DownloadedToolManager.get(project).addAllToPath(this)
        }
    }

    withType<org.curioswitch.gradle.golang.tasks.GoTask>().configureEach {
        if (name.startsWith("goBuild") || name == "goTest") {
            dependsOn(generateProto)
        }
    }

    named("gitPublishReset"){
        doLast {
            val f = File("$buildDir/gitPublish/.git/config")
            if (!f.readText().contains("sys-admin")) {
                f.appendText("\n[user]\nname = InfoStellar Inc\nemail = sys-admin@istellar.jp\n")
            }
        }
    }

    named("gitPublishCopy").configure({
        dependsOn(generateProto, copyGoMod)
    })

    named("assemble").configure({
        dependsOn(generateProto, copyGoMod)
    })

    // Only generated code, no need to check.
    named("goCheck").configure({
        enabled = false
    })
    named("goTest").configure({
        enabled = false
    })
}
