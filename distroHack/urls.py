from django.conf.urls import patterns, url

from distroHack import views


urlpatterns = patterns('',
    # /hacks/
    url(r'^question/(?P<q_id>\d+)/$', views.question, name='question'),
    url(r'^runcode/$', views.runcode, name='runcode'),
    url(r'^updateq/$', views.update_question, name='updateQuestion'),
    url(r'^signin/$', views.sign_in, name='sign_in'),
    url(r'^register/$', views.register, name='register'),
    url(r'^logout/$', views.logout, name='logout'),


    # /polls/
    url(r'^polls/$', views.polls_index, name='polls_index'),
    url(r'^polls/(?P<poll_id>\d+)/$', views.polls_detail, name='polls_detail'),
    url(r'^polls/(?P<poll_id>\d+)/results/$', views.polls_results, name='polls_results'),
    url(r'^polls/(?P<poll_id>\d+)/vote/$', views.polls_vote, name='polls_vote'),
)
