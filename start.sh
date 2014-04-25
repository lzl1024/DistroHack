#export GOPATH=$1/goServer

#go build serverdb
#go build superNode
go build mainServer


python manage.py syncdb

if [ "$#" -eq 3  ]
then
  $1/mainServer $2 $3 &
  python manage.py runserver $3:8000
else
  $1/mainServer $2 &
  python manage.py runserver 8000
fi