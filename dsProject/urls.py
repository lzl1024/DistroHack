from django.conf.urls import patterns, include, url

from django.contrib import admin
admin.autodiscover()

urlpatterns = patterns('',
    (r'^$', 'distroHack.views.index'),
    (r'^(?P<poll_id>\d+)/$', 'distroHack.views.detail'),
    (r'^(?P<poll_id>\d+)/results/$', 'distroHack.views.results'),
    (r'^polls/(?P<poll_id>\d+)/vote/$', 'distroHack.views.vote'),
    url(r'^admin/', include(admin.site.urls)),
)
