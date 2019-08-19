#!/bin/bash

# All parameter fields are required for the script to execute
#
# parameter ${1} is autodeploy
# parameter ${2} is the openshift-master-url
# parameter ${3} is the user-token
#
# parameters 4 & 5 are only needed if ${1} == "true"
# parameter ${4} is the project/namespace in openshift where the deploymentconfig resides
# parameter ${5} is the name of the deploymentconfig for rollout of the new image
#
# example usage : build-deploy.sh true https://master.balt1.okd.14west.io:8443 EI_L1xrzGnOdsJwoCfJ3PrkDNc9OLxliYE4hGuTKG10 test poc-microservice

# declare some variables
name="kaniko-myportfolio-message-producer"

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

if [ -z ${OC_TOKEN} ]; 
then
  echo -e "\033[0;91mOC_TOKEN envar is not set please set it in the environments tab (secure envar) in GOCD\033[0m"
  exit -1
fi

if [ "${AUTODEPLOY}" == "true" ];
then
  if [ -z ${OC_NAMESPACE} ]; 
  then
    echo -e "\033[0;91mOC_NAMESPACE envar is not set please set it in the environments tab in GOCD\033[0m"
    exit -1
  fi
  if [ -z ${OC_DEPLOYMENTCONFIG} ]; 
  then
    echo -e "\033[0;91mOC_DEPLOYMENTCONFIG envar is not set please set it in the environments tab in GOCD\033[0m"
    exit -1
  fi
fi


# list some gocd variables
echo -e " "
echo "GOCD job name         ${GO_JOB_NAME}"
echo "GOCD pipeline name    ${GO_PIPELINE_NAME}"
echo "GOCD pipeline counter ${GO_PIPELINE_COUNTER}"
echo "GOCD pipeline label   ${GO_PIPELINE_LABEL}"
echo -e " " 

# first login
oc login ${MASTER_URL} --username=${OC_USER} --password=${OC_PASSWORD} --insecure-skip-tls-verify -n ci-cd


# we can now execute the job
oc create -f kaniko-job.yml

status=""
while [ "${status}" == "" ]
do
  status=$(oc get job/${name} -o=jsonpath='{.status.conditions[*].type}')
done

pod=$(oc get pods | grep "${name}" | awk '{print $1}')
oc logs po/"${pod}"

if [ "${status}" != "Complete" ];
then
  echo "Failed"
 	exit -1
else
  echo "Passed"
  # if we aren't deploying then just exit
  if [ "${AUTODEPLOY}" == "false" ];
  then
    oc delete job/"${name}"
    exit 0
  fi
fi

# delete the job
oc delete job/"${name}" 

if [ "${AUTODEPLOY}" == "true" ];
then
  # we assume that the project resides on the same server (master-url)
  # if not then add a new login call here first
  oc project ${OC_NAMESPACE}
  oc rollout ${OC_DEPLOYMENTCONFIG}
  exit 0
fi
