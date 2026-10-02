Got sent this on login:

```
CLIENT_PROTOCOL 1
CLIENT_TYPE PS2 ID=0000000000000000
LOGIN fasfsa fasfas
```

In Ghidra I found a `LOGIN_OK` string, sending it successfully auths (client side)
and the clients goes to the chat window. 

It then responds `CHAT`, probably asking for a listing of the available chats?
```
LOGIN_OK
CHAT
```

Reading further in Ghidra, acceptable responses to the `LOGIN` command are `LOGIN_OK` and `LOGIN_FAILED`
