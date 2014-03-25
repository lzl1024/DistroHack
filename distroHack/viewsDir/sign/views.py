from django.shortcuts import render
from django.views.decorators.csrf import csrf_exempt
import re
from distroHack.views import default_tuple, local_ranking


@csrf_exempt
def sign_in(request):
    name = request.POST["name"]
    password = request.POST["password"]
    validate = "failed"
    result_msg = []

    if not name or not password:
        result_msg.append("UserName or Password should not be empty!")
    else:
        #TODO: connect with the go server
        user = []

        # authenticate success, login
        if user is not None:
            validate = "success"
            request.session['username'] = name
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

    # locally check valid information
    if not email or not re.match('.+@.+', email):
        result_msg.append("Email is invalid!")

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

    # create a new user and update database
    if len(result_msg) == 0:

        #TODO: Connect to go server
        msg = []
        #   result_msg.append("This Username has already been registered!")
        #   "This email has already been registered!"

        # remote error
        if msg is not None and len(msg) != 0:
            result_msg.append(msg)
        else:
            validate = "success"
            # login and fill with the local_ranking
            request.session['username'] = name
            user_tuple = default_tuple.copy()
            user_tuple['name'] = name
            local_ranking[name] = user_tuple

    return render(request, 'hack/error.html', {'error': result_msg, 'msg': validate})


# user logout
def logout(request):
    del request.session['username']
    return render(request, 'index.html')
