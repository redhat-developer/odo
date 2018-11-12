import numpy

def application(environ, start_response):
  start_response('200 OK', [('Content-Type','text/plain')])
  matrix = numpy.array([[1,2,3],[6,5,4],[7,8,8]])
  matrix.dot(numpy.linalg.inv(matrix))
  return [b"Hello World from numpy WSGI application!"]
