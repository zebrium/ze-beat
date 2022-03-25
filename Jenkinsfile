@Library('cicd') _
def pipeLineYaml
def buildID
def oldVersion
def newVersion
def repos
pipeline {
    agent any
    stages {
        stage('Prepare Environment') {
            steps {
                script {
                    pipeLineYaml = readYaml(file: "${WORKSPACE}/pipeline.yaml")
                    repos = readYaml(text: libraryResource("repositories.yaml"))
                }
            }
        }
        stage ('Set Project Version') {
            steps {
                script {
                    buildtime = env.BUILD_TIMESTAMP_CI
                    oldVersion = sh (script: "sed -n '1p' version.txt", returnStdout: true).trim()
                    newVersion = oldVersion
                    //TODO: Increment this based on tagged version
                    newVersion = sh (script: "echo ${newVersion} | awk -F. -v OFS=. '{\$2++;\$NF=0;print}'", returnStdout:  true).trim()
                    echo "Updating Version of Plugin"
                    sh """sed -i 's/${oldVersion}/${newVersion}/g' version.txt"""
                }
                buildName "${newVersion}"
            }
        }
        stage("Checking Code Coverage & Unit Testing"){
            steps{
                script {
                    sonarScanner(newVersion)
                }
            }
        }
         stage('Build Docker Image') {
            steps {
                script {
                   sh "aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 872295030327.dkr.ecr.us-west-2.amazonaws.com"
                   sh "docker build -t ${pipeLineYaml.id}:${newVersion} --network host . --no-cache"
                }
            }
        }
        stage('Publish Docker Image') {
            steps {
                publishDockerImages(newVersion, false, pipeLineYaml.id, repos)
            }
        }
        stage ('Push new Version'){
            steps{
                script{
                    echo "Publishing New version to bitbucket"
                    sh """git config user.email "jenkins@zebrium.com" && git config user.name "Jenkins" && git commit -am "Updated version to ${newVersion} from ${oldVersion}" """
                    sshagent(['bitbucket']) {
                        sh 'git push --set-upstream origin $BRANCH_NAME'
                    }
                }
            }
        }
    }
    post {
        always {
            cleanWs()
        }
    }
}