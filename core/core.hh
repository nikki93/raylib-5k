#pragma once // Doubled-up header guards to satisfy 'clangd'
#ifndef __CORE_H__
#define __CORE_H__


#include "core/precomp.hh"
#include "core/meta.hh"

//#define GX_NO_CHECKS
#define GX_FIELD_ATTRIBS PropAttribs
#include "gx.hh"
template<typename T, unsigned N = 0>
using Seq = gx::Slice<T>;

#include "core/misc.hh"
#include "core/entity.hh"
#include "core/rl.hh"
#include "core/c2.h"
#include "core/read_write.hh"
#include "core/ui.hh"
#include "core/edit.hh"


#endif
