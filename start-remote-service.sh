creds_path="config/auth.json"
gcloud auth activate-service-account backend@pokernotifications-229105.iam.gserviceaccount.com --key-file=${creds_path} --project=pokernotifications-229105
gcloud compute ssh --strict-host-key-checking=no --zone=us-central1-c colerwyats@hopper-instance-1 --command="sudo systemctl start bot.service"