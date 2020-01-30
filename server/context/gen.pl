#!/usr/bin/perl

use strict;
use warnings;

use Data::Dumper;

my $struct = [
	'config', '*common.Configuration', { panic => 1 },
	'logger', '*logger.Logger', {  panic => 1 },

	'metadataBackend', 'metadata.Backend', {  panic => 1  },
	'dataBackend', 'data.Backend', {  panic => 1  },
	'streamBackend', 'data.Backend', {  panic => 1 },

	'sourceIp', 'net.IP', {},

	'upload', '*common.Upload', {},
	'file', '*common.File', {},
	'user', '*common.User', {},
	'token', '*common.Token', {},

	'isWhitelisted', 'bool', {},
	'isAdmin', 'bool', {},
	'isUploadAdmin', 'bool', {},
	'isRedirectOnFailure', 'bool', {},
	'isQuick', 'bool', {},
	'isPanic', 'bool', {},

	'req', '*http.Request', {},
	'resp', 'http.ResponseWriter', {},
];

my @struct_extra = qw(
	goctx gocontext.Context
	mu sync.RWMutex
);

sub genHas
{
    my $param = shift;
    my $params = shift;

    return "" if $params->{'no has'};

    my $uc = ucfirst $param;

    my $str = << "EOF";
// Has$uc return true if $param is set for the context
func (ctx *Context) Has$uc() bool {
    ctx.mu.RLock()
    ctx.mu.RUnlock()

    if ctx.$param != nil {
        return true
    }
    return false
}
EOF
    return $str;
}

sub genWith
{
    my $param = shift;
    my $type = shift;
    my $params = shift;

    return "" if defined $params->{'no with'};

    my $uc = ucfirst $param;

    my $str = << "EOF";
// WithConfig set $param for the context
func (ctx *Context) With$uc($param $type) *Context {
    ctx.mu.Lock()
    ctx.mu.Unlock()

    if ctx.$param == nil {
        ctx.$param = $param
    } else {
EOF

    if ( defined $params->{'panic'} )
    {
        $str .= "        ctx.isPanic = true";
    }

    $str .= << "EOF";
        ctx.InternalServerError(internalServerError, fmt.Errorf("context $param overwrite"))
    }
    return ctx
}
EOF

    return $str;
}

sub genGet
{
    my $param = shift;
    my $type = shift;
    my $params = shift;

    return "" if defined $params->{'no get'};

    my $uc = ucfirst $param;

    my $str = << "EOF";
// GetConfig from the request context.
func (ctx *Context) Get$uc() (config $type) {
    ctx.mu.RLock()
    ctx.mu.RUnlock()

    if ctx.$param == nil {
EOF

    if ( defined $params->{'panic'} )
    {
        $str .= "   ctx.isPanic = true";
    }

    $str .= << "EOF";
        ctx.InternalServerError(internalServerError, fmt.Errorf("missing context $param"))
    }
    return ctx.$param
}
EOF

}

for (my $i = 0 ; $i < @$struct ; $i += 3)
{
    my $param = $struct->[$i];
    my $type = $struct->[$i + 1];
    my $params = $struct->[$i + 2];

    my $uc = ucfirst $param;


    print genHas($param, $params);
    print genWith($param, $type, $params);
    print genGet($param, $type, $params);
}