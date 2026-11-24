return function(reader)

	local entries = {}

	while true do
		local len = reader:read4()
		if len == 0xFFFFFFFF then
			break
		end
		local str = reader:readBytes(len)
		entries[#entries + 1] = str
	end

	return table.concat(entries, "\n")
end