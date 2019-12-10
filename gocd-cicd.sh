#!/bin/bash

# All parameter fields are required for the script to execute

# declare some variables
PROJECT="trackmate-message-producer"
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
echo "PROJECT                 ${PROJECT}"
echo "SONARQUBE_HOST          ${SONARQUBE_HOST}"
echo "SONARQUBE_SCANNER_PATH  ${SONARQUBE_SCANNER_PATH}"
echo "GOCD job name           ${GO_JOB_NAME}"
echo "GOCD pipeline name      ${GO_PIPELINE_NAME}"
echo "GOCD pipeline counter   ${GO_PIPELINE_COUNTER}"
echo "GOCD pipeline label     ${GO_PIPELINE_LABEL}"
echo -e " " 

if [ "$1" = "sonarqube" ]
then
   echo -e "\nSonarqube scanning project"
   rm -rf output.json
   touch output.json
   fs=$(stat --printf='%s\n' output.json)
   result="\"PENDING\""
   ${SONARQUBE_SCANNER_PATH}bin/sonar-scanner -Dsonar.projectKey=${PROJECT} -Dsonar.sources=. -Dsonar.host.url=${SONARQUBE_HOST} -Dsonar.login=${SONARQUBE_USER} -Dsonar.password=${SONARQUBE_PASSWORD} -Dsonar.go.coverage.reportPaths=tests/results/cover.out -Dsonar.exclusions=vendor/**,*_test.go,main.go,connectors.go,schema.go,swaggerui/**,tests/**,*.json,*.txt,*.yml,*.xml,*.sh,Dockerfile -Dsonar.issuesReport.json.enable=true -Dsonar.report.export.path=sonar-report.json -Dsonar.issuesReport.console.enable=true | tee response.txt
   # response text includes the url to view the json payload of the sonar scanner results
   url=$(cat response.txt | grep -o "${SONARQUBE_HOST}/api/ce/task?id=[A-Za-z0-9\-]*")
   # loop until we have a valid payload
   while [[ ${fs} -eq 0 ]] && [[ "${result}" = "\"PENDING\"" ]];
   do
     sleep 2;
     curl -H 'Content-Type: application/json' -H 'Accept: application/json' -H 'Authorization: Basic YWRtaW46Yml0bmFtaQ==' "${url}" > output.json;
     fs=$(stat --printf='%s\n' output.json);
     result=$(cat output.json | jq '.task.status');
     echo "${fs} ${result}";
   done
   # check to see if the job was succesful
   echo ${result} | grep -o "SUCCESS" && echo "PASSED" && exit 0 || echo "FAILED" && exit 1
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
