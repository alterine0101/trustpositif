# database pemblokiran [Trust+] dalam satu regex;
aka. the whole [Trust+] database in one regular expression (#_ );

> ![NOTE]
> we recommend to use the **reverse-domain detection** for best performance: reverse the domain text (e.g. `example.com` &rarr; `moc.elpmaxe`) them match them with the one found in `output/regex-reversed.txt`;
>
> the original `regex.txt` will no longer be tested due to performance issues;

## motivation
+ i'm bored
+ i'm overpowered

## instruction
download the [`domains`](https://trustpositif.kominfo.go.id/assets/db/domains) file && leave it as is under the `/input` subdirectory;

run `main.go` program && 

a bunch of regex will be generated under the `/output` subdirectory;

## behind the scenes
this program generates a freakin' huge [trie](https://en.wikipedia.org/wiki/Trie) tree then compiles into a bunch of regex; hooray!

## future plans
- [ ] simplify that freakin' large trie to a [DAG](https://en.wikipedia.org/wiki/Directed_acyclic_graph), so we can convert:

```
[a]->[.]-+->[c]->[o]->[m]
         |
         +->[b]->[.]->[c]->[o]->[m]
         |
         +->[.]->[n]->[e]->[t]

Regex: a\.((com)|(b\.com)|(\.net))
```

into this:

```
         +->[b]->[.]
         |          \
[a]->[.]-+-----------+->[c]->[o]->[m]
         |
         +->[.]->[n]->[e]->[t]

Regex: a\.(((b\.){0,1}com)|(\.net))
```

so regex engines can be more efficient in parsing this huge one.

[Trust+]: https://trustpositif.kominfo.go.id
