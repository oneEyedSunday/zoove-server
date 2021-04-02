==> Allow users create tokens as devs.
==> be able to revoke this token. how? by checking if the token has a jwt id and aud. aud is the id of the user/dev who created the key and the jwt id is the id of the token itself. when a request is made to revoke a token, the server checks if it has jwtid, if so, it adds it to the redis-store alongside the aud to know which user owns the key.

==> query redis on every api call and check if the jwt id of blacklisted tokens contain the jwt of the incoming token. if so, tell them token has been invalidated. else, proceed.
