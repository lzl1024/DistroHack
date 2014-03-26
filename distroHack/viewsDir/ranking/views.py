# show current global ranking
from django.shortcuts import render
from distroHack.views import global_ranking, min_show_rank_len, default_tuple


# show the ranking chart
def ranks(request):
    # login control
    if 'username' not in request.session:
        return render(request, 'hack/please_log_in.html')

    length = len(global_ranking)

    # add more element into ranking when its length is too small
    if length < min_show_rank_len:
        for i in range(min_show_rank_len - length):
            global_ranking.append(default_tuple)
        length = min_show_rank_len
    return render(request, 'hack/rank.html', {'rank': global_ranking, 'length': length})


# TODO update global ranking when receive msgs from lower level server
def update_rank(request):
    return None


# TODO update local information when receive msgs from lower level server
def update_local(request):
    pass