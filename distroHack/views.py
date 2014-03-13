import subprocess
import random
import os
from threading import Lock
import urllib

from django.shortcuts import render, get_object_or_404
from django.http import HttpResponseRedirect, HttpResponse, Http404
from django.core.urlresolvers import reverse
from django.shortcuts import render_to_response
from django.views.decorators.csrf import csrf_exempt
import sys
import shutil

from distroHack.models import Poll, Choice
from dsProject.settings import OJ_PATH

# id_set to record the used submitted_id
id_set = set()
lock = Lock()

# judge policy and timeout
policy = "test.policy"
timeout = str(10)
run_script = OJ_PATH + "/run"
source_name = "Source"
result_name = "result.txt"


def polls_vote(request, poll_id):
    p = get_object_or_404(Poll, pk=poll_id)
    try:
        selected_choice = p.choice_set.get(pk=request.POST['choice'])
    except (KeyError, Choice.DoesNotExist):
        # Redisplay the poll voting form.
        return render(request, 'polls/detail.html', {
            'poll': p,
            'error_message': "You didn't select a choice.",
            })
    else:
        selected_choice.votes += 1
        selected_choice.save()
        # Always return an HttpResponseRedirect after successfully dealing
        # with POST data. This prevents data from being posted twice if a
        # user hits the Back button.
        return HttpResponseRedirect(reverse('polls:polls_results', args=(p.id,)))


def polls_index(request):
    latest_poll_list = Poll.objects.order_by('-pub_date')[:5]
    context = {'latest_poll_list': latest_poll_list}
    return render(request, 'polls/index.html', context)


def polls_detail(request, poll_id):
    poll = get_object_or_404(Poll, pk=poll_id)
    return render(request, 'polls/detail.html', {'poll': poll})


def polls_results(request, poll_id):
    poll = get_object_or_404(Poll, pk=poll_id)
    return render(request, 'polls/results.html', {'poll': poll})


def index(request):
    return render_to_response('index.html')


def question(request, q_id):
    return render(request, 'hack/question.html', {'q_id': q_id})


@csrf_exempt
def runcode(request):
    if request.is_ajax():
        # create unique id
        submit_id = create_id()
        if submit_id == -1:
            raise Http404

        print "Submit id: " + str(submit_id)

        # get file path and make dir
        file_dir_path = os.path.join(OJ_PATH, str(submit_id))
        source_path = os.path.join(file_dir_path, source_name)
        result_path = os.path.join(file_dir_path, result_name)

        try:
            os.stat(file_dir_path)
        except os.error:
            os.mkdir(file_dir_path)

        # write source code into file
        source_file = open(source_path + ".java", "wb")
        source_file.write(urllib.unquote(request.POST["code"]))
        source_file.close()

        # run command and wait for completion
        cmd = run_script + " " + file_dir_path + " " + source_name + " " + \
            result_name + " " + policy + " " + timeout
        process = subprocess.Popen(['/bin/sh', '-c', cmd])
        process.wait()

        # read the result
        result_file = open(result_path, "r")
        result_msg = result_file.read()

        # clear the created dir and files
        result_file.close()
        if os.path.exists(file_dir_path):
            shutil.rmtree(file_dir_path)

        return HttpResponse(result_msg)
    else:
        raise Http404


# create unique id and put into id_set
def create_id():
    lock.acquire()
    while 1:
        new_id = random.randint(0, sys.maxint)
        if new_id not in id_set:
            id_set.add(new_id)
            lock.release()
            return new_id
