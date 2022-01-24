label = "${UUID.randomUUID().toString()}"
BUILD_FOLDER = "/home/jenkins/go"
attempts=15
git_project = "v3io-go"
git_project_user = "v3io"
git_project_upstream_user = "v3io"
git_deploy_user = "iguazio-prod-git-user"
git_deploy_user_token = "iguazio-prod-git-user-token"
git_deploy_user_private_key = "iguazio-prod-git-user-private-key"

podTemplate(label: "${git_project}-${label}", inheritFrom: "jnlp-docker-golang") {
//    def MAIN_TAG_VERSION
//    def FRAMES_NEXT_VERSION
//    def next_versions = ['prometheus':null, 'tsdb-nuclio':null, 'frames':null]

    pipelinex = library(identifier: 'pipelinex@development', retriever: modernSCM(
            [$class:        'GitSCMSource',
             credentialsId: git_deploy_user_private_key,
             remote:        "git@github.com:iguazio/pipelinex.git"])).com.iguazio.pipelinex

    prnumber_split="${JOB_NAME.substring(JOB_NAME.lastIndexOf('/') + 1, JOB_NAME.length()).toLowerCase()}"
    prnumber="${prnumber_split.substring(prnumber_split.lastIndexOf('-') + 1, prnumber_split.length()).toLowerCase()}"

    common.main {
        node("${git_project}-${label}") {
            withCredentials([
                    string(credentialsId: git_deploy_user_token, variable: 'GIT_TOKEN')
            ]) {
                stage('get tag data') {
//                    container('jnlp') {
//                        MAIN_TAG_VERSION = github.get_tag_version(TAG_NAME)
                        sh """
                        pwd
                        ls -ltrh 
                        """
//                        echo "$MAIN_TAG_VERSION"
//                    }
                }
            }
        }
//        node("${git_project}-${label}") {
//            withCredentials([
//                    string(credentialsId: git_deploy_user_token, variable: 'GIT_TOKEN')
//            ]) {
//                stage('update release status') {
//                    container('jnlp') {
//                        github.update_release_status(git_project, git_project_user, "${MAIN_TAG_VERSION}", GIT_TOKEN)
//                    }
//                }
//            }
//        }
    }
}
