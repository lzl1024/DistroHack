from django.conf.urls import patterns, include, url

from django.contrib import admin
admin.autodiscover()

urlpatterns = patterns('',
    url(r'^$', 'distroHack.views.index'),
    url(r'^hacks/', include("distroHack.urls", namespace="distroHack")),
    url(r'^admin/', include(admin.site.urls)),
)
