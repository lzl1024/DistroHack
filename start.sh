
export GOPATH=$1/goServer
go build mainServer

$1/mainServer &

python manage.py runserver 8000
