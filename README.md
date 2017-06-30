# cryptchat
simple terminal based secure chat client

[![codecov](https://codecov.io/gh/bryandmc/cryptchat/branch/master/graph/badge.svg)](https://codecov.io/gh/bryandmc/cryptchat)  [![Build Status](https://travis-ci.org/bryandmc/cryptchat.svg?branch=master)](https://travis-ci.org/bryandmc/cryptchat)

The idea is to have a simple chat client that will use PGP to encrypt a message on your end and decrypt on another user's end with a pgp key shared 'out of band.' (over email or in person or wherever). The backend will use a simple redis instance to store / distribute messages. This is not p2p so chat 'rooms' will be possible. This is just for fun. Maybe the API will be further developed to allow the core functionality to be used for others to build on top of (maybe). 

