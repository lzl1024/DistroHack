from django.conf.urls import patterns, url

from distroHack import views


urlpatterns = patterns('',
    # /hacks/
    url(r'^question/$', 'distroHack.viewsDir.question.views.question', name='question'),
    url(r'^runcode/$', 'distroHack.viewsDir.question.views.runcode', name='runcode'),
    url(r'^updateq/$', 'distroHack.viewsDir.question.views.update_question', name='updateQuestion'),
    url(r'^signin/$', 'distroHack.viewsDir.sign.views.sign_in', name='sign_in'),
    url(r'^register/$', 'distroHack.viewsDir.sign.views.register', name='register'),
    url(r'^logout/$', 'distroHack.viewsDir.sign.views.logout', name='logout'),
    url(r'^ranks/$', 'distroHack.viewsDir.ranking.views.ranks', name='ranks'),
    url(r'^update_rank/$', 'distroHack.viewsDir.ranking.views.update_rank', name='update_rank'),
    url(r'^update_local/$', 'distroHack.viewsDir.ranking.views.update_local', name='update_local'),
    url(r'^admin/$', 'distroHack.views.admin', name='admin'),
    url(r'^start_hack/$', 'distroHack.views.start_hack', name='start_hack'),
    url(r'^end_hack/$', 'distroHack.views.end_hack', name='end_hack'),
)
