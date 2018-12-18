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
    id("org.curioswitch.gradle-curio-web-plugin")
}

web {
    javaPackage.set("com.stellarstation.developers.portal")
}

tasks {
    val mergePortal by registering(Copy::class) {
        dependsOn(":stubs:python:generateDocs", "buildWeb")

        into("build/site")

        from("build/web")

        from("${project(":stubs:python").buildDir}/docs/stellarstation") {
            into("python/apidocs")
        }
    }

    val deployRelease by registering(org.curioswitch.gradle.plugins.nodejs.tasks.NodeTask::class) {
        dependsOn(mergePortal)

        args("run", "deploy-release")
    }
}
