@Library('cicd') _
def pipeLineYaml
def buildID
def semver
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
                    def oldVersion = sh (script: "sed -n '1p' version.txt", returnStdout: true).trim()
                    newVersion = oldVersion
                    //TODO: Increment this based on tagged version
                    newVersion = sh (script: "echo ${newVersion} | awk -F. -v OFS=. '{\$2++;\$NF=0;print}'", returnStdout:  true).trim()
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
                   sh "docker build -t ${pipeLineYaml.id}:${tag} --network host . --no-cache"
                }
            }
        }
        stage('Publish Docker Image') {
            steps {
                publishDockerImages(tag, false, pipeLineYaml.id, repos)
            }
        }
    }
    post {
        always {
            cleanWs()
        }
    }
}