pipeline {
    agent {
        node {
            label 'mesos-ubuntu'
        }
    }
    stages {
        stage('test') {
            steps {
                sh 'make test'
            }
        }
    }
}
