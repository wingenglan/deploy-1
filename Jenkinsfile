pipeline {
  agent any

  parameters {
    choice(
      name: 'PROJECT',
      choices: ['all', 'frontend', 'backend'],
      description: '要处理的项目：all=前后端都处理，frontend=仅前端，backend=仅后端'
    )
    choice(
      name: 'BRANCH',
      choices: ['main', 'dev'],
      description: '要拉取构建的分支'
    )
    booleanParam(
      name: 'DEPLOY',
      defaultValue: false,
      description: '是否执行服务器部署（勾选后才会部署到服务器，务必确认分支正确）'
    )
  }

  environment {
    DEPLOY_PATH = '/opt/flowboard'
  }

  stages {
    stage('Checkout') {
      steps {
        // 拉取指定分支的代码；github-token 是上一步在凭据里创建的 ID
        checkout([
          $class: 'GitSCM',
          branches: [[name: "${params.BRANCH}"]],
          userRemoteConfigs: [[
            url: 'https://github.com/wingenglan/deploy-1.git',
            credentialsId: 'github-token'
          ]]
        ])
      }
    }

    stage('Verify backend') {
      when { expression { params.PROJECT in ['all', 'backend'] } }
      steps {
        dir('backend') {
          sh 'go test ./...'
        }
      }
    }

    stage('Build frontend') {
      when { expression { params.PROJECT in ['all', 'frontend'] } }
      steps {
        dir('frontend') {
          sh 'npm ci'
          sh 'npm run build'
        }
      }
    }

    stage('Deploy') {
      when { expression { params.DEPLOY } }
      steps {
        withCredentials([sshUserPrivateKey(
          credentialsId: 'deploy-ssh-key',
          keyFileVariable: 'SSH_KEY',
          usernameVariable: 'SSH_USER'
        )]) {
          sh '''
            # 在服务器上拉取所选分支的最新代码，再按所选项目执行部署脚本
            ssh -i ${SSH_KEY} -o StrictHostKeyChecking=no \
              ${SSH_USER}@${DEPLOY_HOST} \
              "cd ${DEPLOY_PATH} && \
               git fetch origin ${BRANCH} && \
               git reset --hard origin/${BRANCH} && \
               sh scripts/deploy.sh ${PROJECT}"
          '''
        }
      }
    }
  }
}
