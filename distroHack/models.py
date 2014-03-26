from django.db import models


class Problem(models.Model):
    title = models.CharField(max_length=200)
    description = models.TextField()
    startCode = models.TextField()
    testCode = models.TextField()
    result = models.TextField()
