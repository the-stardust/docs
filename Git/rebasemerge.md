## merge
官方解释：Join two or more development histories together

所以说当你想让两个分支进行汇总的时候，应该使用merge

## rebase
rebase其实就是修改本分支的branch out的点

## 实例解析
```
     A---B---C topic
    /         \
D---E---F---G---H master
```

- 当我们从E切换开发分支A的时候，此时master分支也在往前走，A分支也往前走，此时想吧topic合并到master的时候，就有两种选择了
- git merge ,用merge代表了topic分支与master分支交汇，并解决了所有合并冲突。然而merge的缺点是引入了一次不必要的history join。如图：
```
     A--B--C-X topic
    /       / \
D---E---F---G---H master
```
- 其实仔细想一下就会发现，在引入master分支的F、G commit这个问题上，我们并没有要求两个分支必须进行交汇(join)，我们只是想避免最终的merge conflict而已。
- rebase是另一个选项。rebase的含义是改变当前分支branch out的位置。这个时候进行rebase其实意味着，将topic分支branch out的位置从E改为G，如图：
```
              A--B--C topic
             / 
D---E---F---G---H master
```
- 在这个过程中会解决引入F、G导致的冲突，同时没有多余的history join。但是rebase的缺点是，改变了当前分支branch out的节点

## 总结

综上，其实选用merge还是rebase取决于你到底是以什么意图来避免merge conflict。实践上个人还是偏爱rebase。

!> 一个是因为branch out节点不能改变的情况实在太少。
另外就是频繁从master merge导致的冗余的history join会提高所有人的认知成本。

