#export GOPATH=$1/goServer

#go build serverdb
#go build superNode
go build mainServer

if [ "$#" -eq 3  ]
then
  $1/mainServer $2 $3 &
else
  $1/mainServer $2 &
fi


python manage.py syncdb

python manage.py runserver 8000