pipeline {
  agent any

  environment {
    DEPLOY_PATH = '/opt/flowboard'
  }

  stages {
    stage('Verify backend') {
      steps {
        dir('backend') {
          sh 'go test ./...'
        }
      }
    }

    stage('Build frontend') {
      steps {
        dir('frontend') {
          sh 'npm ci'
          sh 'npm run build'
        }
      }
    }

    stage('Deploy') {
      when { branch 'main' }
      steps {
        sshagent(credentials: ['deploy-ssh-key']) {
          sh '''
            ssh -o StrictHostKeyChecking=no -p ${DEPLOY_PORT:-22} ${DEPLOY_USER}@${DEPLOY_HOST} \
              "mkdir -p ${DEPLOY_PATH}"
            rsync -az --delete -e "ssh -o StrictHostKeyChecking=no -p ${DEPLOY_PORT:-22}" \
              --exclude '.git' --exclude 'frontend/node_modules' ./ ${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}/
            ssh -o StrictHostKeyChecking=no -p ${DEPLOY_PORT:-22} ${DEPLOY_USER}@${DEPLOY_HOST} \
              "cd ${DEPLOY_PATH} && sh scripts/deploy.sh"
          '''
        }
      }
    }
  }
}
