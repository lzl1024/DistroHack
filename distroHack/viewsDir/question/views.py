from threading import Lock
import datetime
import random
import os
import urllib
from django.http import Http404, HttpResponse
from django.shortcuts import render
from django.views.decorators.csrf import csrf_exempt
import subprocess
import shutil
from distroHack.models import Problem
from distroHack.views import local_ranking, default_tuple, global_ranking, show_rank_len
from dsProject.settings import OJ_PATH, PRO_PATH
import sys

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


# question page
def question(request):
    if 'username' not in request.session:
        return render(request, 'hack/please_log_in.html')

    # empty database check
    if Problem is None:
        return render(request, 'hack/notready.html')

    question_number = Problem.objects.count()

    if question_number == 0:
        return render(request, 'hack/notready.html')

    login_user = request.session['username']
    qid = 1
    if local_ranking.get(login_user) is not None:
        # noinspection PyTypeChecker
        qid = local_ranking[login_user]['score'] + 1

        if question_number < qid:
            return render(request, 'hack/complete.html')
    else:
        local_ranking[login_user] = default_tuple.copy()
        local_ranking[login_user]['name'] = login_user

    problem = Problem.objects.get(pk=qid)
    return render(request, 'hack/question.html', {'problem': problem})


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

    return render(request, 'index.html')


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
        problem_id = int(request.POST["id"])
        user = request.session["username"]

        try:
            os.stat(file_dir_path)
        except os.error:
            os.mkdir(file_dir_path)

        # get model
        problem = Problem.objects.get(pk=problem_id)

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
        if problem.result.strip() == result_msg.strip():
            result_msg = "Accepted"

            # add user into local ranking if he is not
            if local_ranking.get(user) is None:
                user_tuple = default_tuple.copy()
                user_tuple['name'] = user
                local_ranking[user] = user_tuple

            # update the ranking locally
            if local_ranking[user]['score'] < problem_id:
                local_ranking[user]['score'] = problem_id
                local_ranking[user]['time'] = datetime.datetime.now()

                # check user in global ranking
                user_index = -1
                for i in range(len(global_ranking)):
                    if global_ranking[i]['name'] == user:
                        user_index = i

                # update old global ranking
                if user_index > 0:
                    global_ranking[user_index]['score'] = problem_id
                elif problem_id > global_ranking[-1]['score'] or len(global_ranking) < show_rank_len:
                    global_ranking.append(local_ranking[user])

                # sort global ranking
                global_ranking.sort(key=lambda k: (k['name'], k['time']), reverse=True)

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