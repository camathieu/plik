#!/usr/bin/perl

use strict;
use warnings;

use Data::Dumper;

my $struct = [
	'config', '*common.Configuration', { 'panic' => 1 },
	'logger', '*logger.Logger', { 'panic' => 1 },

	'metadataBackend', 'metadata.Backend', { 'panic' => 1 },
	'dataBackend', 'data.Backend', { 'panic' => 1 },
	'streamBackend', 'data.Backend', { 'panic' => 1 },

	'sourceIP', 'net.IP', {},

	'upload', '*common.Upload', {},
	'file', '*common.File', {},
	'user', '*common.User', {},
	'token', '*common.Token', {},

	'isWhitelisted', 'bool', {},
	'isAdmin', 'bool', {},
	'isUploadAdmin', 'bool', {},
	'isRedirectOnFailure', 'bool', {},
	'isQuick', 'bool', {},

	'req', '*http.Request', {},
	'resp', 'http.ResponseWriter', {},

    'panic', 'bool', { 'internal' => 1 },
	'mu', 'sync.RWMutex', { 'internal' => 1 },
];

#sub genHas
#{
#    my $param = shift;
#    my $type = shift;
#    my $params = shift;
#
#    return "" if $type eq 'bool';
#    return "" if $params->{'no_has'};
#
#    my $uc = ucfirst $param;
#
#    my $str = << "EOF";
#// Has$uc return true if $param is set in the context
#func (ctx *Context) Has$uc() bool {
#    ctx.mu.RLock()
#    defer ctx.mu.RUnlock()
#
#    if ctx.$param != nil {
#        return true
#    }
#
#    return false
#}
#
#EOF
#    return $str;
#}

sub genGet
{
    my $param = shift;
    my $type = shift;
    my $params = shift;

    return "" if $params->{'internal'};

    my $uc = ucfirst $param;

    my $str = "";
    if ( $type eq 'bool' ) {
        $str = << "EOF";
// $uc get $param from the context.
func (ctx *Context) $uc() $type {
    ctx.mu.RLock()
    defer ctx.mu.RUnlock()

    return ctx.$param
}

EOF
    } else {
        $str = << "EOF";
// Get$uc get $param from the context.
func (ctx *Context) Get$uc() $type {
    ctx.mu.RLock()
    defer ctx.mu.RUnlock()

EOF
        if ( $params->{'panic'} ) {
        $str .= << "EOF";
    if ctx.$param == nil {
        ctx.internalServerError("missing $param from context", nil)
    }

EOF
        }

    $str .= << "EOF";
    return ctx.$param
}

EOF
    }
}

sub genSet
{
    my $param = shift;
    my $type = shift;
    my $params = shift;

    return "" if $params->{'internal'};

    my $uc = ucfirst $param;

    if ( $type eq 'bool' ) {
        $uc =~ s/^Is//
    }

    my $str = << "EOF";
// Set$uc set $param in the context
func (ctx *Context) Set$uc($param $type) {
    ctx.mu.Lock()
    ctx.mu.Unlock()

    ctx.$param = $param
}

EOF

    return $str;
}

sub genImports
{
    my $str = << 'EOF';
package context

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	//gocontext "context"
	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
	"github.com/root-gg/plik/server/metadata"
)

EOF
    return $str;
}

sub genStruct
{
    my $struct = shift;

    my $str = "type Context struct {\n";
    for (my $i = 0 ; $i < @$struct ; $i += 3)
    {
        my $param = $struct->[$i];
        my $type = $struct->[$i + 1];

        $str .= "\t$param $type\n";
    }
    $str .= "}\n";

    return $str;
}

sub genMethods
{
    my $struct = shift;

    my $str = "";
    for (my $i = 0 ; $i < @$struct ; $i += 3)
    {
        my $param = $struct->[$i];
        my $type = $struct->[$i + 1];
        my $params = $struct->[$i + 2];

        my $uc = ucfirst $param;

        $str .= genGet($param, $type, $params);
        $str .= genSet($param, $type, $params);
    }

    return $str;
}

sub gen
{
    print genImports;
    print "\n";
    print genStruct $struct;
    print "\n";
    print genMethods $struct;
    print "\n";
}

gen;