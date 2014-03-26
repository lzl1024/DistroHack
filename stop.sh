
ps aux | grep $1/mainServer | awk '{print $2}' | xargs kill

rm $1/mainServer 


