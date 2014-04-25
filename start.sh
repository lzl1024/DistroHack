#export GOPATH=$1/goServer

#go build serverdb
#go build superNode
go build mainServer

$1/mainServer $2 $3 &

python manage.py syncdb

python manage.py runserver $3:8000
