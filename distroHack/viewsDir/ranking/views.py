# show current global ranking
import json
from dateutil import parser
from django.shortcuts import render
from django.views.decorators.csrf import csrf_exempt
from distroHack.views import min_show_rank_len, default_tuple
import distroHack.views


# show the ranking chart
def ranks(request):
    # login control
    if 'username' not in request.session:
        return render(request, 'hack/please_log_in.html')

    length = len(distroHack.views.global_ranking)

    # add more element into ranking when its length is too small
    if length < min_show_rank_len:
        for i in range(min_show_rank_len - length):
            distroHack.views.global_ranking.append(default_tuple)
        length = min_show_rank_len
    return render(request, 'hack/rank.html',
                  {'rank': distroHack.views.global_ranking, 'length': length})


@csrf_exempt
def update_rank(request):
    data = json.loads(request.POST['data'])

    # parse the data to new global_ranking
    distroHack.views.global_ranking = []
    for element in data:
        new_tuple = default_tuple.copy()
        new_tuple['name'] = element['UserName']
        new_tuple['score'] = element['Score']
        new_tuple['time'] = parser.parse(element['Ctime'])
        distroHack.views.global_ranking.append(new_tuple)

    return render(request, 'index.html')


@csrf_exempt
def update_local(request):
    data = json.loads(request.POST['data'])

    # parse the data to new local_ranking
    distroHack.views.local_ranking = {}
    for key, value in data.iteritems():
        new_tuple = default_tuple.copy()
        new_tuple['name'] = value['UserName']
        new_tuple['score'] = value['Score']
        new_tuple['time'] = parser.parse(value['Ctime'])
        distroHack.views.local_ranking[key] = new_tuple

    return render(request, 'index.html')