#!/bin/bash

{
  cd $(git rev-parse --show-toplevel)
  echo '> SHOWING ALL TERRAFORM FILES'
  sleep 2
} 2> /dev/null
tail ./terraform/**/*.tf

{
  sleep 5
  echo -e '\n> BUILDING AND RUNNING TERRASYNC ON BACKGROUND'
  sleep 3
} 2> /dev/null
bash -cx 'make &' 2> /dev/null

{
  sleep 3
  echo -e '\n> MAKING AN HTTP REQUEST TO TERRASYNC SERVER'
  sleep 3
} 2> /dev/null
curl localhost:8080

{
  sleep 5
  echo -e '\n> terraform/3 IS NOT OUT OF SYNC, LETS CHANGE THIS'
  sleep 3
} 2> /dev/null
sed -i 's/100/300/' terraform/3/main.tf

{
  sleep 3
  echo -e '\n> WEVE CHANGED THE FILE, SO THE SERVER WILL SHOW OUT OF SYNC'
  sleep 3
} 2> /dev/null
git diff

{
  sleep 3
  echo -e '\n> terraform/3 SHOULD BE OUT OF SYNC NOW, MAKING ANOTHER HTTP REQUEST'
  sleep 3
} 2> /dev/null
curl localhost:8080

{
  git restore terraform/3/main.tf
  pkill terrasync > /dev/null
  sleep 5
} 2> /dev/null
