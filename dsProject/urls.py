from django.conf.urls import patterns, include, url

from django.contrib import admin
admin.autodiscover()

urlpatterns = patterns('',
    url(r'^polls/', include("distroHack.urls", namespace="distroHack")),
    url(r'^$', include("distroHack.urls", namespace="distroHack")),
    url(r'^admin/', include(admin.site.urls)),
)
