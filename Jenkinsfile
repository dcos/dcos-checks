pipeline {
    agent {
        node {
            label 'mesos-ubuntu'
        }
    }
    stages {
        stage('testing') {
            steps {
			    sh 'make test'
            }
        }
    }
}
