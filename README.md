An Distributed Hackthon Application.


Having fun with that...


To run this project, you need "mysql" with a database named "test", javac compiler and python/django.
you may run it as normal django project, or import into PyCharm and run.


Before running it: please run "python manage.py syncdb" to synchronize database.

After running it: please first request "http://127.0.0.1:8000/hacks/updateq/" to update the question. Otherwise, you may get a "not yet ready" page when you click on online judge page. Problems are stored in ~/problems/{{id}}/ with test code, start code, result and its description. Online Judge scripts are stored in ~/oj/ repo.
