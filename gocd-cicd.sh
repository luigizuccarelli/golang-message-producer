#!/bin/bash

# All parameter fields are required for the script to execute

# declare some variables
project="trackmate-message-producerr"
jobname="kaniko-trackmate-message-producer"
deploymentconfig="trackmate-message-producer.json"
namespace="myportfolio"


# some variable checks
if [ -z ${MASTER_URL} ]; 
then
  echo -e "\033[0;91mMASTER_URL envar is not set please set it in the environments tab in GOCD\033[0m"
  exit -1
fi

if [ -z ${AUTODEPLOY} ]; 
then
  echo -e "\033[0;913mAUTODEPLOY envar is not set please set it in the environments tab in GOCD\033[0m"
  exit -1
fi

if [ -z ${OC_USERNAME} ]; 
then
  echo -e "\033[0;91mOC_TOKEN envar is not set please set it in the environments tab (secure envar) in GOCD\033[0m"
  exit -1
fi

# list some gocd variables
echo -e " "
echo "GOCD job name         ${GO_JOB_NAME}"
echo "GOCD pipeline name    ${GO_PIPELINE_NAME}"
echo "GOCD pipeline counter ${GO_PIPELINE_COUNTER}"
echo "GOCD pipeline label   ${GO_PIPELINE_LABEL}"
echo -e " " 

if [ "$1" = "sonarqube" ]
then
    echo -e "\nSonarqube scanning project"
    /sonarqube/bin/sonar-scanner -Dsonar.projectKey=${project} -Dsonar.sources=. -Dsonar.host.url=${SONARQUBE_HOST} -Dsonar.login=${SONARQUBE_USER} -Dsonar.password=${SONARQUBE_PASSWORD} -Dsonar.go.coverage.reportPaths=tests/results/cover.out -Dsonar.exclusions=vendor/**,*_test.go,main.go,connectors.go,schema.go,tests/**,*.json,*.txt,*.yml,*.xml,*.sh,Dockerfile -Dsonar.issuesReport.json.enable=true -Dsonar.report.export.path=sonar-report.json -Dsonar.issuesReport.console.enable=true | tee output.txt && \
    result=$(cat output.txt | grep -o "INFO: EXECUTION SUCCESS") &&  echo ${result} | grep -o 'INFO: EXECUTION SUCCESS' && echo "PASSED" && exit 0 || echo "FAILED" && exit -1
fi

if [ "$1" = "build-image" ]
then
  # first login
  oc login ${MASTER_URL} --username=${OC_USER} --password=${OC_PASSWORD} --insecure-skip-tls-verify -n cicd

  # we can now execute the job
  oc create -f kaniko-job.yml

  status=""
  while [ "${status}" == "" ]
  do
    status=$(oc get job/${jobname} -o=jsonpath='{.status.conditions[*].type}')
  done

  pod=$(oc get pods | grep "${jobname}" | awk '{print $1}')
  oc logs po/"${pod}"

  if [ "${status}" != "Complete" ];
  then
    echo "Failed"
 	  exit -1
  else
    echo "Passed"
    # if we arent deploying then just exit
    if [ "${AUTODEPLOY}" == "false" ];
    then
      oc delete job/"${jobname}"
      exit 0
    fi
  fi

  # delete the job
  oc delete job/"${jobname}" 

  if [ "${AUTODEPLOY}" == "true" ];
  then
    # we assume that the project resides on the same server (master-url)
    # if not then add a new login call here first
    oc project ${namespace}
    oc rollout latest dc/${deploymentconfig}
    exit 0
  fi
fi
