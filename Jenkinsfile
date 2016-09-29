#!groovy

node {
	// Lowercase version of the job name
	def name = env.BUILD_TAG.toLowerCase()

	// get the uid of the user running the job to be able to properly manage permissions
	def parentUID = sh(script: 'id -u', returnStdout: true).trim()

	// use docker gid to give job access to docker socket
	def parentGID = sh(script: 'getent group docker | cut -d: -f3', returnStdout: true).trim()

	try {
		stage 'Checkout and cleanup'

			checkout scm

			// Remove ignored files
			sh 'git clean -X -f'

			// Remove code coverage files (not always deleted by git clean if a directory is removed)
			sh "find . -name cover.out -exec rm '{}' \\;"
			sh "find . -name coverage.xml -exec rm '{}' \\;"

			def gitBranch = env.BRANCH_NAME
			def gitCommit = sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
			echo "${gitBranch}"
			echo "${gitCommit}"

			def deploy = false
			if (gitBranch == "master") {
				echo "DEPLOYING"
				deploy = true
			} else {
				echo "NOT DEPLOYING"
			}

			def deployToS3 = ""
			if (deploy) {
				deployToS3 = "true"
			}
			env.FULLCOVERAGE = ""
			env.TEST_S3_BUCKET = ""
			env.NO_INTEGRATION_TESTS = "true"
			env.NO_COVERAGE = "true"

			def memPath = "/mnt/mem/jenkins/" + name
			sh "mkdir -p ${memPath}"

		stage 'Build test docker image'

			sh "docker build --rm --force-rm -t ${name} docker-ci"
			sh "docker pull 137987605457.dkr.ecr.us-east-1.amazonaws.com/scratch:latest"

		stage 'Build and Test'

			def workspace = pwd()
			sh """docker run --rm=true --name=${name} \
                -e "BUILD_NUMBER=${env.BUILD_NUMBER}" \
                -e "BUILD_ID=${env.BUILD_ID}" \
                -e "BUILD_URL=${env.BUILD_URL}" \
                -e "BUILD_TAG=${env.BUILD_TAG}" \
                -e "GIT_COMMIT=${gitCommit}" \
                -e "GIT_BRANCH=${gitBranch}" \
                -e "JOB_NAME=${env.JOB_NAME}" \
                -e "DEPLOY_TO_S3=${deployToS3}" \
                -e "FULLCOVERAGE=${env.FULLCOVERAGE}" \
                -e "TEST_S3_BUCKET=${env.TEST_S3_BUCKET}" \
                -e "PARENT_UID=${parentUID}" \
                -e "PARENT_GID=${parentGID}" \
                -e "NO_INTEGRATION_TESTS=${env.NO_INTEGRATION_TESTS}" \
                -e "NO_COVERAGE=${env.NO_COVERAGE}" \
                -v ${memPath}:/mem \
                -v ${workspace}:/workspace/go/src/github.com/sprucehealth/backend \
                -v /var/run/docker.sock:/var/run/docker.sock \
                ${name}"""

		stage 'Deploy'

			if (deploy) {
				env.GIT_REV = gitCommit
				env.BRANCH = gitBranch
				sh(script: "./docker-ci/deploy.sh")
			}
	} catch (any) {
		currentBuild.result = 'FAILURE'
		throw any //rethrow exception to prevent the build from proceeding
	} finally {
		step([$class: 'Mailer', notifyEveryUnstableBuild: true, recipients: 'backend@sprucehealth.com', sendToIndividuals: true])
		step([$class: 'JUnitResultArchiver', testResults: '**/tests.xml'])
	}
}