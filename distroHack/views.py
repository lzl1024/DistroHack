import datetime
from django.shortcuts import render


# ranking constants
default_time = datetime.datetime(2014, 1, 31, 8, 31, 24)
global_ranking = [{'name': 'lzl', 'score': 4, 'time': default_time}]
local_ranking = {'lzl': {'name': 'lzl', 'score': 4, 'time': default_time}}
default_tuple = {'name': '', 'score': 0, 'time': ''}
show_rank_len = 20
min_show_rank_len = 3


# main page
def index(request):
    return render(request, 'index.html')

