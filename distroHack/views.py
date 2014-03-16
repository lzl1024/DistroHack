import subprocess
import random
import os
import django
from threading import Lock
import urllib
from django.contrib.auth import authenticate, login
from django.contrib.auth.models import User
from django.shortcuts import render, get_object_or_404, redirect
from django.http import HttpResponseRedirect, HttpResponse, Http404
from django.core.urlresolvers import reverse
from django.shortcuts import render_to_response
from django.views.decorators.csrf import csrf_exempt
import sys
import shutil
import re

from distroHack.models import Poll, Choice, Problem
from dsProject.settings import OJ_PATH, PRO_PATH

# id_set to record the used submitted_id
id_set = set()
lock = Lock()

# judge policy and timeout
policy = "test.policy"
timeout = "10"
run_script = OJ_PATH + "/verify"
source_name = "Source"
result_name = "result.txt"
test_name = "Test"
error_flag = "Error"

# problem files
descript_file_name = "description.txt"
answer_file_name = "answer.txt"
start_file_name = "startCode.java"
test_file_name = "testCode.java"


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
        test_path = os.path.join(file_dir_path, test_name)

        try:
            os.stat(file_dir_path)
        except os.error:
            os.mkdir(file_dir_path)

        # get model
        problem = Problem.objects.get(pk=int(request.POST["id"]))

        # write source/test code into file
        source_file = open(source_path + ".java", "wb")
        source_file.write(urllib.unquote(request.POST["code"]))
        source_file.close()
        test_file = open(test_path + ".java", "wb")
        test_file.write(problem.testCode.encode('utf-8'))
        test_file.close()

        # run command and wait for completion
        cmd = run_script + " " + file_dir_path + " " + source_name + " " + \
              test_name + " " + result_name + " " + policy + " " + timeout
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

        # judge accept or denied
        if problem.result == result_msg:
            result_msg = "Accepted"
            #TODO if accept, send msg to lower level server
            
        elif result_msg.strip().endswith(error_flag):
            result_msg = "Denied\n" + result_msg
        else:
            result_msg = "Denied"

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
    descpt_file = open(os.path.join(problem_dir, descript_file_name), "rb")
    problem.title = descpt_file.readline()
    problem.description = descpt_file.read()
    descpt_file.close()

    # other fields
    problem.result = read_fields(problem_dir, answer_file_name)
    problem.startCode = read_fields(problem_dir, start_file_name)
    problem.testCode = read_fields(problem_dir, test_file_name)
    problem.save()

    return redirect('distroHack.views.index')


@csrf_exempt
def sign_in(request):
    name = request.POST["name"]
    password = request.POST["password"]
    validate = "failed"
    result_msg = []

    if not name or not password:
        result_msg.append("UserName or Password should not be empty!")
    else:
        user = authenticate(username=name, password=password)
        # authenticate success, login
        if user is not None:
            validate = "success"
            login(request, user)
        else:
            result_msg.append("Name-Password pair is not registered!")

    return render(request, 'hack/error.html', {'error': result_msg, 'msg': validate})


@csrf_exempt
def register(request):
    # pick and initiate fields
    email = request.POST["email"]
    name = request.POST["userName"]
    password = request.POST["password"]
    com_password = request.POST["confirm"]
    result_msg = []
    validate = "failed"

    # check valid information
    if not email or not re.match('.+@.+', email):
        result_msg.append("Email is invalid!")
    elif User.objects.filter(email=email):
        result_msg.append("This email has already been registered!")

    if not password or not com_password:
        result_msg.append("Password and Confirm Password should not be empty!")
    else:
        if password != com_password:
            result_msg.append("Confirm Password is not the same as the Password!")
        elif len(password) not in range(5, 21):
            result_msg.append("The length of the password should between 5 and 20!")

    if not name:
        result_msg.append("UserName should not be empty!")
    elif len(name) not in range(1, 201):
        result_msg.append("The length of the UserName should between 1 and 200!")
    elif User.objects.filter(username=name):
        result_msg.append("This Username has already been registered!")

    # create a new user and update database
    if len(result_msg) == 0:
        user = User.objects.create_user(username=name,
                                    password=password,
                                    email=email)
        user.is_active = True
        user.save()
        validate = "success"
        login(request, user)

    return render(request, 'hack/error.html', {'error': result_msg, 'msg': validate})


def logout(request):
    django.contrib.auth.logout(request)
    return redirect('distroHack.views.index')