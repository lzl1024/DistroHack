import subprocess
import random
import os
from threading import Lock
import urllib

from django.shortcuts import render, get_object_or_404, redirect
from django.http import HttpResponseRedirect, HttpResponse, Http404
from django.core.urlresolvers import reverse
from django.shortcuts import render_to_response
from django.views.decorators.csrf import csrf_exempt
import sys
import shutil

from distroHack.models import Poll, Choice, Problem
from dsProject.settings import OJ_PATH, PRO_PATH

# id_set to record the used submitted_id
id_set = set()
lock = Lock()

# judge policy and timeout
policy = "test.policy"
timeout = "10"
run_script = OJ_PATH + "/run"
source_name = "Source"
result_name = "result.txt"

# problem files
descript_name = "description.txt"
answer_name = "answer.txt"
start_name = "startCode.java"
test_name = "testCode.java"


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


# main page
def index(request):
    return render_to_response('index.html')


# question page
def question(request, q_id):
    if Problem is None or Problem.objects.count() < int(q_id):
        return render(request, 'hack/notready.html')

    problem = Problem.objects.get(pk=int(q_id))
    return render(request, 'hack/question.html', {'problem': problem})


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
        lock.acquire()
        if os.path.exists(file_dir_path):
            shutil.rmtree(file_dir_path)
        id_set.remove(submit_id)
        lock.release()

        #TODO if accept, send msg to lower level server

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


# simple tool function to read fields in files
def read_fields(problem_dir, field_name):
    tmp_file = open(os.path.join(problem_dir, field_name), "rb")
    result = tmp_file.read()
    tmp_file.close()
    return result


# update the database and fill with problems
def update_question(request):
    p_id = 1
    problem = Problem.objects.get_or_create(id=1)[0]
    problem_dir = os.path.join(PRO_PATH, str(p_id))

    # title and description
    descpt_file = open(os.path.join(problem_dir, descript_name), "rb")
    problem.title = descpt_file.readline()
    problem.description = descpt_file.read()
    descpt_file.close()

    # other fields
    problem.result = read_fields(problem_dir, answer_name)
    problem.startCode = read_fields(problem_dir, start_name)
    problem.testCode = read_fields(problem_dir, test_name)
    problem.save()

    return redirect('distroHack.views.index')