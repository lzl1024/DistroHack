import datetime
import json
from django.shortcuts import render
from django.views.decorators.csrf import csrf_exempt
# ranking constants
from distroHack.viewsDir.sign.views import connect_server

default_time = datetime.datetime(2014, 1, 31, 8, 31, 24)
global_ranking = [{'name': 'lzl', 'score': 4, 'time': default_time}]
default_tuple = {'name': '', 'score': 0, 'time': ''}
show_rank_len = 20
min_show_rank_len = 3

# start and end time
hack_end_time = None
hack_is_started = False


# main page
def index(request):
    return render(request, 'index.html')


# admin page
def admin(request):
    if request.session["username"] == "admin":
        return render(request, 'admin.html', {"time": hack_end_time})

    return render(request, 'index.html')


# admin to start the hackathon
@csrf_exempt
def start_hack(request):
    global hack_end_time, hack_is_started
    if 'data' in request.POST:
        # get info from others
        data = json.loads(request.POST['data'])
        hack_end_time = datetime.datetime(int(data['year']), int(data['month']),
                                      int(data['day']), int(data['hour']),
                                      int(data['minute']), int(data['second']))
    else:
        hack_end_time = datetime.datetime(int(request.POST['year']), int(request.POST['month']),
                                      int(request.POST['day']), int(request.POST['hour']),
                                      int(request.POST['minute']), int(request.POST['second']))
    hack_is_started = True

    # only admin connect to server
    if request.session["username"] == "admin":
        # dump to json format and send out
        msg = {"type": "start_hack", "year": request.POST['year'], "month": request.POST['month'],
        "day": request.POST['day'], "hour": request.POST['hour'], "minute": request.POST['minute'],
        "second": request.POST['second']}
        connect_server(msg)

    return render(request, 'index.html')


# admin to end the hackathon
@csrf_exempt
def end_hack(request):
    global hack_is_started, hack_end_time
    hack_is_started = False
    hack_end_time = None

    # only admin connect to server
    if request.session["username"] == "admin":
    # dump to json format and send out
        msg = {"type": "end_hack"}
        connect_server(msg)

    return render(request, 'index.html')
