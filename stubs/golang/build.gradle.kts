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
        artifact.set("com.google.protobuf:protoc:3.6.1")
    }

    // Don"t use descriptor set.
    descriptorSetOptions.path.set(file("build/descriptor"))

    languages {
        register("go") {
            option("plugins=grpc")
            plugin {
                path.set(file("${GOPATH}/bin/protoc-gen-go"))
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

        args("install", "github.com/golang/protobuf/protoc-gen-go")
    }

    val installProtoWrap by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        dependsOn(named("goDownloadDeps"))

        args("install", "github.com/square/goprotowrap/cmd/protowrap")
    }

    val installMockGen by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        dependsOn(named("goDownloadDeps"))

        args("install", "github.com/golang/mock/mockgen")
    }

    val copyGoMod by registering(Copy::class) {
        into("build/generated/proto/main/github.com/infostellarinc/go-stellarstation")
        from("go.mod")
        from("go.sum")
    }

    val runModVendoring by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        args("mod", "vendor")
        dependsOn(copyGoMod)

        execCustomizer {
            workingDir = file("build/generated/proto/main/github.com/infostellarinc/go-stellarstation")
        }
    }

    val generateProto by getting(org.curioswitch.gradle.protobuf.tasks.GenerateProtoTask::class) {
        dependsOn(installProtocGoPlugin, installProtoWrap)
        finalizedBy(copyGoMod, runModVendoring)

        execOverride {
            val protowrapPath = project.file("${GOPATH}/bin/protowrap")
            setCommandLine(listOf(protowrapPath.getAbsolutePath(), "--protoc_command=${executable}") + args)

            org.curioswitch.gradle.tooldownloader.DownloadedToolManager.get(project).addAllToPath(this)
        }
    }

    // Because mockgen isn't aware of Go modules, we need to prepare a GOPATH for it like this.
    val copyDepsToMockgenGopath by registering(Copy::class) {
        into("build/goproto/src/")
        dirMode = 493 // 755
        fileMode = 420 // 644

        dependsOn(generateProto, installMockGen, runModVendoring)

        val goModLines = file("go.mod").readLines()
        var foundRequire = false

        for (line in goModLines) {
            if (!foundRequire && !line.startsWith("require")) {
                continue
            }

            if (line.startsWith("require")) {
                foundRequire = true
                continue
            }

            if (line.startsWith(")")) {
                break
            }
            val parts = line.trim().split(" ")
            from("$GOPATH/pkg/mod/${parts[0]}@${parts[1]}"){
                into(parts[0])
                exclude("**/testdata/**")
            }
        }
    }

    val copyBuildedSource by registering(Copy::class) {
        into("build/goproto/src/github.com/infostellarinc/go-stellarstation")
        from("build/generated/proto/main/github.com/infostellarinc/go-stellarstation")
        dirMode = 493 // 755
        fileMode = 420 // 644

        dependsOn(generateProto, installMockGen, runModVendoring)
    }

    val runMockgenStellarStationServiceClient by registering(org.curioswitch.gradle.golang.tasks.GoTask::class) {
        val outputDir = project.file("build/generated/proto/main/github.com/infostellarinc/go-stellarstation/api/mock_v1")
        val outputFile = project.file("${outputDir}/stellarstation.mock.go")

        inputs.dir(project.file("build/generated/proto/main/github.com/infostellarinc/go-stellarstation"))
        outputs.dir(outputDir)

        dependsOn(copyDepsToMockgenGopath, copyBuildedSource)

        command(project.file("${GOPATH}/bin/mockgen").toString())
        args("-destination=${outputFile}".toString(), "github.com/infostellarinc/go-stellarstation/api/v1", "StellarStationServiceClient")

        execCustomizer({
            environment("GOPATH", file("build/goproto").getAbsolutePath())
            environment("GO111MODULE", "off")
        })

    }

    withType<org.curioswitch.gradle.golang.tasks.GoTask>().configureEach {
        if (name.startsWith("goBuild") || name == "goTest") {
            dependsOn(generateProto)
            dependsOn(runMockgenStellarStationServiceClient)
            execCustomizer({
                environment("GOFLAGS", "-mod=vendor")
            })
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
        dependsOn(generateProto, runMockgenStellarStationServiceClient)
    })

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

    named("compileJava") {
        enabled = false
    }

    named("compileTestJava") {
        enabled = false
    }
}
