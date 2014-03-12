from django.conf.urls import patterns, url

from distroHack import views

urlpatterns = patterns('',


    # /polls/
    url(r'^polls/$', views.polls_index, name='polls_index'),
    url(r'^polls/(?P<poll_id>\d+)/$', views.polls_detail, name='polls_detail'),
    url(r'^polls/(?P<poll_id>\d+)/results/$', views.polls_results, name='polls_results'),
    url(r'^polls/(?P<poll_id>\d+)/vote/$', views.polls_vote, name='polls_vote'),
)
