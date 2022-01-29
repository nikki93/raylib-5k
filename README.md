<p float="left" align="center">
  <img src="readme_header.png" height="250">
</p>

# Packet Lost

This is the code and assets repository for our (me, Luka and Axi) Raylib 5K jam submission, ["Packet Lost"](https://itch.io/jam/raylib-5k-gamejam/rate/1374384)!

It uses my Go->C++ transpiler, called [Gx](https://github.com/nikki93/gx), along with my own little framework that includes an entity system, a scene editor and seriaization. The framework doesn't wrap Raylib at all--Raylib is used directly. The jam entry doesn't use the serialization system much since the world is procedurally generated, but the editor came in handy during development to debug and explore the resulting scene. Interesting things to look at:

- [game/behaviors.gx.go](game/behaviors.gx.go): This file lists the components that can be attached to an entity, along with some other data structures. This basically shows the structure of the game's state.
- [game/game.gx.go](game/behaviors.gx.go): This composes the entirety of the game's runtime logic, over the above data. It's all in one file because we sped through this in the jam--usually I would split it up into files meant for each aspect of the game. No header files needed since Gx's module system (coming from Go) just makes things work.

The files under [core/](core/) make up the engine. [core/entity.hh](core/entity.hh) is the entity system implementation, which could be cool to look at if you're curious how the data is stored. [core/read_write.hh](core/read_write.hh) implements the serialization system. All the game art and sound go in [assets/](assets/).

I'll be adding much more information here soon! Part of the motivation for doing the jam was to have a resulting open source example like this that uses Gx. But the jam was tiring and now I should probably rest for a bit so... Will get to that in a little while... 😅
