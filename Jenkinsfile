@Library('pipelinex@development') _

def props = [
        buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '30', numToKeepStr: '1000')),
        disableConcurrentBuilds()
]

properties([
        buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '5', numToKeepStr: '10')),
        parameters([
                string(name: 'system_id', defaultValue: '',
                        description: 'Name that consist of lower case alphanumeric chars, "-" or "." and must start and end with an alphanumeric char', trim: true)

        ]),

])

sys_conf = ""

timestamps {
    common.main {
        node('centos76_runner') {
            prnumber_split = "${JOB_NAME.substring(JOB_NAME.lastIndexOf('/') + 1, JOB_NAME.length()).toLowerCase()}"
            prnumber = "${prnumber_split.substring(prnumber_split.lastIndexOf('-') + 1, prnumber_split.length()).toLowerCase()}"
            try {
                if (env.BUILD_NUMBER == "1") {
                    system_id = "${JOB_NAME.substring(JOB_NAME.lastIndexOf('/') + 1, JOB_NAME.length()).toLowerCase()}-${BUILD_NUMBER}"
                    common.conditional_stage('deploy eks', true) {
                        common.run_job("customer_deploy_aws/development", [
                                string(name: 'system_id', value: system_id),
                                booleanParam(name: 'deploy', value: true),
                                booleanParam(name: 'delay_delete', value: false),
                                booleanParam(name: 'proceed_deletion', value: false),
                                string(name: 'version', value: "3.2.0-b19.20211107120205"),
                                string(name: 'kompton_branch', value: "development"),
                                string(name: 'data_nodes_count', value: "1"),
                                string(name: 'app_nodes_count', value: "1"),
                                string(name: 'frontend_nodes_count', value: "0"),
                                string(name: 'data_nodes_type', value: "i3.4xlarge"),
                                string(name: 'app_nodes_type', value: "m5.4xlarge"),
                                string(name: 'frontend_nodes_type', value: "m5.4xlarge"),
                                string(name: 'provazio_vpc_mode', value: "existing"),
                                string(name: 'provazio_public_ip_kind', value: "static"),
                                string(name: 'encryption_type', value: "none"),
                                string(name: 'duration', value: "8"),
                                string(name: 'data_node_root_volume_size', value: "400"),
                                string(name: 'app_node_root_volume_size', value: "400"),
                                string(name: 'provazio_version', value: "latest"),
                                string(name: 'api_key', value: "provctl_api_key_customer_deploy_aws"),
                                string(name: 'vault_env', value: "dev"),
                                string(name: 'provazio_github_user', value: "iguazio"),
                                string(name: 'provazio_dashboard_repo', value: "gcr.io/iguazio/provazio-dashboard"),
                                booleanParam(name: 'sanity', value: false),
                                booleanParam(name: 'db_report', value: true),
                                booleanParam(name: 'log_export', value: true),
                                booleanParam(name: 'copy_data_sets', value: true),
                                string(name: 'naipi_branch', value: "development"),
                                string(name: 'naipi_params', value: "--cycle-definition=IG-1734"),
                                string(name: 'tf_log_level', value: "DEBUG"),
                                string(name: 'k8s_type', value: "iguazio"),
                                string(name: 'eks_lifecycle', value: "ondemand"),
                                string(name: 'cni_kind', value: "calico"),
                                booleanParam(name: 'limit_services_to_initial_node_group', value: true),
                                string(name: 'secondary_app_nodes_count', value: "0"),
                                string(name: 'secondary_app_nodes_type', value: "m5.4xlarge"),
                                string(name: 'eks_min_second_nodes_group', value: "0"),
                                string(name: 'eks_max_second_nodes_group', value: "0"),

                        ])
                    }
                }

            }
            catch (err) {
                echo err.getMessage()
                echo "will continue to next stage"
            }
//                currentBuild.result='FAILURE'
//            } finally {
//                    echo "delete sysytem:  ${system_id}"
////                    stages.delete_system(system_id)
//            }

            stage('git clone') {
                deleteDir()
                def scm_vars = checkout scm
                env.git_hash = scm_vars.GIT_COMMIT
                currentBuild.description = "hash = ${env.git_hash}"
            }

//            stage('set_env') {
//
//                def sys_conf = sh(script: "http --verify no --check-status -b GET http://dashboard.dev.provazio.iguazio.com/api/systems/${system_id}",returnStdout: true)
//                system_config = readJSON(text: sys_conf)
//                env.V3IO_DATAPLANE_URL = system_config['status']['tenants'][1]['status']['services']['webapi']['api_urls']['https']
//                env.V3IO_DATAPLANE_USERNAME = system_config['spec']['tenants'][1]['spec']['resources'][0]['users'][0]['username']
//                env.V3IO_CONTROLPLANE_URL = system_config['status']['tenants'][1]['status']['services']['dashboard']['urls']['https']
//                env.V3IO_CONTROLPLANE_USERNAME = system_config['spec']['tenants'][1]['spec']['resources'][0]['users'][0]['username']
//                env.V3IO_CONTROLPLANE_PASSWORD = system_config['spec']['tenants'][1]['spec']['resources'][0]['users'][0]['password']
//                env.V3IO_DATAPLANE_ACCESS_KEY = sh(script: "./hack/script/generate_access_key.sh", returnStdout: true).split('=')[1].trim()
//
//            }


//
//
//
            stage('build') {

                def sys_conf = sh(script: "http --verify no --check-status -b GET http://dashboard.dev.provazio.iguazio.com/api/systems/${system_id}", returnStdout: true)
                system_config = readJSON(text: sys_conf)
                env.V3IO_DATAPLANE_URL = system_config['status']['tenants'][1]['status']['services']['webapi']['api_urls']['https']
                env.V3IO_DATAPLANE_USERNAME = system_config['spec']['tenants'][1]['spec']['resources'][0]['users'][0]['username']
                env.V3IO_CONTROLPLANE_URL = system_config['status']['tenants'][1]['status']['services']['dashboard']['urls']['https']
                env.V3IO_CONTROLPLANE_USERNAME = system_config['spec']['tenants'][1]['spec']['resources'][0]['users'][0]['username']
                env.V3IO_CONTROLPLANE_PASSWORD = system_config['spec']['tenants'][1]['spec']['resources'][0]['users'][0]['password']
                env.V3IO_DATAPLANE_ACCESS_KEY = sh(script: "./hack/script/generate_access_key.sh", returnStdout: true).split('=')[1].trim()

                    withCredentials([
                            usernamePassword(credentialsId: 'igz_admin', usernameVariable: 'V3IO_CONTROLPLANE_IGZ_ADMIN_USERNAME', passwordVariable: 'V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD'),
                    ]) {
                        try {

                        sh """
                echo "testting"
                make test-system-in-docker
                """
                    } catch (err) {
                        echo err.getMessage()
                        echo "will continue to next stage"
                    } finally {
                        echo "delete sysytem:  ${system_id}"
                         stages.delete_system(system_id)
                    }


                }
            }
//
//
//
//            stage('run Test') {
//                sh "ls -ltrh "
//
//            }

//            stage('Git Merge') {
//               stages.git_merge_pr(system_id,"iguazio","devops-functions")
//
//            }
//
//            stage('Delete  system') {
//
//            }


        }


    }


}

