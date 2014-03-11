from django.shortcuts import render

# Create your views here.
from django.shortcuts import render_to_response


def vote(request, poll_id):
    return render_to_response('vote.html')


def index(request):
    return render_to_response('index.html')


def detail(request, poll_id):
    return render_to_response('index.html')


def results(request, poll_id):
    return render_to_response('index.html')
