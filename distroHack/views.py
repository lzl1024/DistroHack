import datetime
from django.shortcuts import render

# ranking constants
default_time = datetime.datetime(2014, 1, 31, 8, 31, 24)
global_ranking = [{'name': 'lzl', 'score': 4, 'time': default_time}]
local_ranking = {'lzl': {'name': 'lzl', 'score': 4, 'time': default_time}}
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
def start_hack(request):
    if request.session["username"] == "admin":
        global hack_end_time, hack_is_started
        hack_end_time = datetime.datetime(int(request.POST['year']), int(request.POST['month']),
                                        int(request.POST['day']), int(request.POST['hour']),
                                        int(request.POST['minute']), int(request.POST['second']))
        hack_is_started = True
    return render(request, 'index.html')


# admin to end the hackathon
def end_hack(request):
    print "aa"
    if request.session["username"] == "admin":
        global hack_is_started, hack_end_time
        hack_is_started = False
        hack_end_time = None

    return render(request, 'index.html')
