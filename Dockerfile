FROM iron/base

RUN mkdir /app
WORKDIR /app
ADD dequeuer /app
CMD /app/dequeuer
