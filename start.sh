export GOPATH=$1/goServer
go build mainServer

$1/mainServer $2 &

python manage.py runserver 8000
