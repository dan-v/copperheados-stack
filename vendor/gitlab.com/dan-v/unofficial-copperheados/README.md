## Pre-requisites
* Install Go (1.9+)
* AWS admin credentials

## Running
  ```sh
  go run main.go -n copperheados-dan -r us-west-2 --ssh-key macbook13 -d marlin
  ```

## To Do
* Verify end to end build is working
* Restrict created IAM roles to minimum required privileges (currently all admin)
* Add Cloudwatch event for calling Lambda S3 cleanup function