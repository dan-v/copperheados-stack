<b>copperheados-stack</b> is an tool that will deploy all the infrastructure required to run your own [CopperheadOS](https://copperhead.co/android/) build and release environment in AWS. It uses AWS Lambda to check for new releases, provisions spot instances for OS builds, and uploads build artifacts to S3. Resulting OS builds are configured to receive OTA updates from this environment.

## Features
* Support for Google Pixel and Pixel XL (additional supported devices can be added)
* End to end setup of build environment for CopperheadOS in AWS
* Scheduled Lambda function looks for updated builds on a daily basis
* OTA updates through built in updater app - no need to manually flash your device on each new release
* Costs just a few dollars a month to run (EC2 spot instance and S3 storage costs)

## Supporting CopperheadOS
If you use CopperheadOS, I <b>HIGHLY</b> recommend supporting the project with donations: https://copperhead.co/android/donate. 

## Installation
The easiest way is to download a pre-built binary from the [GitHub Releases](https://github.com/dan-v/copperheados-stack/releases) page.

## Prerequisites
You'll need AWS CLI credentials setup with 'AdministratorAccess': https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html

## Deployment
* Deploy environment for Pixel XL in AWS (marlin)

    ```sh
    ./copperheados-stack --region us-west-2 --name copperheados-dan --device marlin
    ```

* Deploy environment for Pixel in AWS (sailfish)

    ```sh
    ./copperheados-stack --region us-west-2 --name copperheados-dan --device sailfish
    ```

* Remove environment and all AWS resources

    ```sh
    ./copperheados-stack --remove --region us-west-2 --name copperheados-dan
    ```

## First Time Setup After Deployment
* Go to the AWS Lambda console and execute the function \<stackname>-build to kick off first build (or you could wait for daily cron to kick it off). This build will take a few hours to complete.
* After build finishes, a factory image should be uploaded to the S3 bucket '\<stackname>-release'. From this bucket, download the file '\<device>-factory-\<build_date>.tar.xz'. 
* Use this factory image and follow the instructions on flashing your device: https://copperhead.co/android/docs/install
* After successful flash, your device will now have CopperheadOS on it and be able to perform OTA updates going forward.

## FAQ
1. <b>Should I use copperheados-stack?</b> That's up to you. Use at your own risk.

# Powered by
* [Terraform](https://www.terraform.io/) 

## Build From Source

  ```sh
  make
  ```

## To Do
* Restrict created IAM roles to minimum required privileges (currently all admin)